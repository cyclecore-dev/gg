package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func handleSearch() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg search <query>")
		fmt.Println()
		fmt.Println("Searches npm, PyPI, crates.io, and Go proxy simultaneously.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg search tokio")
		fmt.Println("  gg search express")
		fmt.Println("  gg search requests")
		return
	}
	query := os.Args[2]
	encodedQuery := url.QueryEscape(query)

	fmt.Printf("Search: \"%s\"\n", query)

	npmResults, npmErr := searchNPM(encodedQuery)
	pypiResults, pypiErr := searchPyPI(encodedQuery)
	cratesResults, cratesErr := searchCrates(encodedQuery)
	goResults, goErr := searchGo(query)

	totalResults := len(npmResults) + len(pypiResults) + len(cratesResults) + len(goResults)

	printRegistries(npmResults, npmErr, "npm")
	printRegistries(pypiResults, pypiErr, "pypi")
	printRegistries(cratesResults, cratesErr, "crates.io")
	printRegistries(goResults, goErr, "go")

	fmt.Printf("Token cost: ~%d\n", totalResults*3)
}

func printRegistries(results []RegistryResult, err error, name string) {
	if err != nil {
		fmt.Printf("%s: error\n", name)
		return
	}
	if len(results) == 0 {
		fmt.Printf("%s: no match\n", name)
		return
	}
	fmt.Printf("%s (%d):\n", name, len(results))
	for _, r := range results {
		desc := r.Description
		if len(desc) > 60 {
			desc = desc[:60]
		}
		fmt.Printf("  %s %s — %s\n", r.Name, r.Version, desc)
	}
}

type RegistryResult struct {
	Name        string
	Version     string
	Description string
}

func searchHTTP(urlStr string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gg-cli (https://github.com/ggdotdev/gg)")
	return client.Do(req)
}

func searchNPM(encodedQuery string) ([]RegistryResult, error) {
	urlStr := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=%s&size=3", encodedQuery)
	resp, err := searchHTTP(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Objects []struct {
			Package struct {
				Name        string `json:"name"`
				Version     string `json:"version"`
				Description string `json:"description"`
			} `json:"package"`
		} `json:"objects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []RegistryResult
	for _, obj := range data.Objects {
		results = append(results, RegistryResult{
			Name:        obj.Package.Name,
			Version:     obj.Package.Version,
			Description: obj.Package.Description,
		})
	}
	return results, nil
}

func searchPyPI(encodedQuery string) ([]RegistryResult, error) {
	urlStr := fmt.Sprintf("https://pypi.org/pypi/%s/json", encodedQuery)
	resp, err := searchHTTP(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return []RegistryResult{}, nil
	}

	var data struct {
		Info struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Summary string `json:"summary"`
		} `json:"info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return []RegistryResult{{
		Name:        data.Info.Name,
		Version:     data.Info.Version,
		Description: data.Info.Summary,
	}}, nil
}

func searchCrates(encodedQuery string) ([]RegistryResult, error) {
	urlStr := fmt.Sprintf("https://crates.io/api/v1/crates?q=%s&per_page=3", encodedQuery)
	resp, err := searchHTTP(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Crates []struct {
			Name        string `json:"name"`
			MaxVersion  string `json:"max_version"`
			Description string `json:"description"`
		} `json:"crates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []RegistryResult
	for _, c := range data.Crates {
		results = append(results, RegistryResult{
			Name:        c.Name,
			Version:     c.MaxVersion,
			Description: c.Description,
		})
	}
	return results, nil
}

func searchGo(query string) ([]RegistryResult, error) {
	urlStr := fmt.Sprintf("https://proxy.golang.org/%s/@latest", query)
	resp, err := searchHTTP(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode == 410 {
		return []RegistryResult{}, nil
	}

	var data struct {
		Version string `json:"Version"`
		Time    string `json:"Time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return []RegistryResult{{
		Name:        query,
		Version:     data.Version,
		Description: "",
	}}, nil
}
