package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func callAPIStreaming(cfg *Config, prompt, repo string) (string, error) {
	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)

	if provider != ProviderOllama && apiKey == "" {
		return "", fmt.Errorf("API key not configured. Run: gg config init")
	}

	systemPrompt := fmt.Sprintf("You are a code generation assistant for the repository: %s\n\n"+
		"Generate clean, production-ready code based on the user's request.\n"+
		"Format code blocks as:\n"+
		"```language:path/to/file\n"+
		"code here\n"+
		"```\n\n"+
		"Be concise and only generate the requested code.", repo)

	switch provider {
	case ProviderOpenAI:
		return callOpenAIStreaming(apiKey, model, systemPrompt, prompt)
	case ProviderOllama:
		return callOllamaStreaming(endpoint, model, systemPrompt, prompt)
	default:
		return callAnthropicStreaming(apiKey, model, systemPrompt, prompt, cfg.API.Temperature)
	}
}

func callAnthropicStreaming(apiKey, model, systemPrompt, prompt string, temperature float64) (string, error) {
	if temperature == 0 {
		temperature = 0.7
	}

	requestBody := map[string]interface{}{
		"model":       model,
		"max_tokens":  4096,
		"stream":      true,
		"system":      systemPrompt,
		"temperature": temperature,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", sanitizeError(fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes)))
	}

	// Parse SSE stream
	var fullResponse strings.Builder
	var inputTokens, outputTokens int64
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Usage struct {
				InputTokens  int64 `json:"input_tokens"`
				OutputTokens int64 `json:"output_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			fmt.Print(event.Delta.Text)
			fullResponse.WriteString(event.Delta.Text)
		}

		if event.Type == "message_delta" && event.Usage.OutputTokens > 0 {
			outputTokens = event.Usage.OutputTokens
		}
		if event.Type == "message_start" && event.Usage.InputTokens > 0 {
			inputTokens = event.Usage.InputTokens
		}
	}

	fmt.Println()

	if inputTokens > 0 || outputTokens > 0 {
		trackTokenUsage(inputTokens, outputTokens)
	}

	return fullResponse.String(), nil
}

func callOpenAIStreaming(apiKey, model, systemPrompt, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", sanitizeError(fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes)))
	}

	// Parse SSE stream
	var fullResponse strings.Builder
	var totalTokens int64
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Usage struct {
				TotalTokens int64 `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			fmt.Print(event.Choices[0].Delta.Content)
			fullResponse.WriteString(event.Choices[0].Delta.Content)
		}

		if event.Usage.TotalTokens > 0 {
			totalTokens = event.Usage.TotalTokens
		}
	}

	fmt.Println()

	if totalTokens > 0 {
		// Estimate input/output split (rough 30/70)
		trackTokenUsage(totalTokens*3/10, totalTokens*7/10)
	}

	return fullResponse.String(), nil
}

func callOllamaStreaming(endpoint, model, systemPrompt, prompt string) (string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	requestBody := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 600 * time.Second} // Ollama can be slow
	resp, err := client.Do(req)
	if err != nil {
		return "", sanitizeError(fmt.Errorf("Ollama connection failed: %v (is Ollama running?)", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", sanitizeError(fmt.Errorf("Ollama error (%d): %s", resp.StatusCode, string(bodyBytes)))
	}

	// Ollama returns newline-delimited JSON
	var fullResponse strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		var event struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Message.Content != "" {
			fmt.Print(event.Message.Content)
			fullResponse.WriteString(event.Message.Content)
		}

		if event.Done {
			break
		}
	}

	fmt.Println()

	return fullResponse.String(), nil
}

// ============================================================================
// PACKAGE MANAGER LAYER
// ============================================================================

// handleNPM fetches npm package info and displays MCP endpoint

func callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, userPrompt string) (string, error) {
	switch provider {
	case ProviderAnthropic:
		return callAnthropicWithSystem(apiKey, systemPrompt, userPrompt)
	case ProviderOllama:
		return callOllamaWithSystem(endpoint, "", systemPrompt, userPrompt)
	case ProviderOpenAI:
		return callOpenAIWithSystem(apiKey, systemPrompt, userPrompt)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func callAnthropicWithSystem(apiKey, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 500,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}
	return "", fmt.Errorf("no response content")
}

func callOllamaWithSystem(endpoint, model, systemPrompt, userPrompt string) (string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": userPrompt,
		"system": systemPrompt,
		"stream": false,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(endpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("%s", result.Error)
	}

	return result.Response, nil
}

func callOpenAIWithSystem(apiKey, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": 500,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response content")
}
