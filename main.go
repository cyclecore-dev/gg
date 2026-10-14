// gg v0.10.1 — the 2-letter agent-native git client
// October 2026 — security hardening
package main

import (
	"fmt"
	"os"
	"strings"
)


const version = "0.10.1"

// Backend URL for Pro tier management
const defaultBackendURL = "https://gg-backend.cyclecore.workers.dev"

func getBackendURL() string {
	if url := os.Getenv("GG_BACKEND_URL"); url != "" {
		return url
	}
	return defaultBackendURL
}

// Token costs (approximate)
const (
	TokenCostNPM  = 18
	TokenCostBrew = 22
	TokenCostGit  = 12
	TokenCostBase = 20
)

// Supported providers
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderOllama    = "ollama"
)

// SecretsData holds encrypted API keys
type SecretsData struct {
	APIKey        string `toml:"api_key"`         // Primary API key (any provider)
	ClaudeAPIKey  string `toml:"claude_api_key"`  // Legacy: kept for backwards compat
	MaazaAPIKey   string `toml:"maaza_api_key"`
	ProLicenseKey string `toml:"pro_license_key"`
}

// Config represents the gg configuration
type Config struct {
	GG struct {
		Version string `toml:"version"`
		Tier    string `toml:"tier"`
	} `toml:"gg"`
	API struct {
		Provider    string  `toml:"provider"`    // anthropic, openai, ollama
		Model       string  `toml:"model"`       // Model name
		Temperature float64 `toml:"temperature"`
		Endpoint    string  `toml:"endpoint"`    // Custom endpoint (for Ollama)
		// Legacy fields for backwards compat
		ClaudeModel       string  `toml:"claude_model"`
		ClaudeTemperature float64 `toml:"claude_temperature"`
		MaazaModel        string  `toml:"maaza_model"`
	} `toml:"api"`
	GitHub struct {
		DefaultBranch string `toml:"default_branch"`
	} `toml:"github"`
	Secrets SecretsData `toml:"keys"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("gg v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	case "init":
		initConfig()
	case "config":
		handleConfig()
	case "maaza":
		handleMaaza()
	case ".":
		handleCurrentRepo()
	case "ask":
		handleAsk()
	case "approve":
		handleApprove()
	case "pr":
		handlePR()
	case "run":
		handleRun()
	case "stats":
		handleStats()
	case "edit":
		handleEdit()
	case "prompts":
		handlePrompts()
	case "npm":
		handleNPM()
	case "brew":
		handleBrew()
	case "chain":
		handleChain()
	case "cool":
		handleCool()
	case "cache":
		handleCache()
	case "a2a":
		handleA2A()
	case "chat":
		handleChat()
	case "pip":
		handlePip()
	case "go":
		handleGo()
	case "crate":
		handleCrate()
	case "time":
		handleTime()
	case "snap":
		handleSnap()
	case "diff":
		handleDiff()
	case "log":
		handleLog()
	case "context":
		handleContext()
	case "cat":
		handleCat()
	case "tree":
		handleTree()
	case "search":
		handleSearch()
	case "deps":
		handleDeps()
	case "test":
		handleTest()
	case "reflog":
		handleReflog()
	case "license":
		handleLicense()
	case "push":
		handlePush()
	case "pull":
		handlePull()
	case "cherry-pick":
		handleCherryPick()
	case "revert":
		handleRevert()
	case "maven":
		handleMaven()
	case "docker":
		handleDocker()
	case "pub":
		handlePub()
	case "nuget":
		handleNuget()
	case "fetch":
		handleFetch()
	case "env":
		handleEnv()
	case "add":
		handleAdd()
	case "commit":
		handleCommit()
	case "checkout":
		handleCheckout()
	case "show":
		handleShow()
	case "clean":
		handleClean()
	case "loc":
		handleLoc()
	case "todo":
		handleTodo()
	case "hex":
		handleHex()
	case "stash":
		handleStash()
	case "blame":
		handleBlame()
	case "tag":
		handleTag()
	case "merge":
		handleMerge()
	case "worktree":
		handleWorktree()
	case "reset":
		handleReset()
	case "gem":
		handleGem()
	case "status":
		handleStatus()
	case "branch":
		handleBranch()
	case "clone":
		handleClone()
	case "remote":
		handleRemote()
	case "composer":
		handleComposer()
	case "mcp":
		handleMCP()
	case "upgrade":
		handleUpgrade()
	case "pro":
		handlePro()
	default:
		if strings.Contains(cmd, "/") {
			handleRepo(cmd)
		} else {
			fmt.Printf("gg: unknown command: %s\n", cmd)
			printUsage()
		}
	}
}

func printUsage() {
	fmt.Printf("gg v%s — token compression for AI agents\n", version)
	fmt.Println("Compresses git/npm/pip/brew calls to <100 tokens (vs 1,000-2,000 raw)")
	fmt.Println()
	fmt.Println("quick start:")
	fmt.Println("  gg pip requests      # 18 tokens (vs ~1,200 raw)")
	fmt.Println("  gg npm lodash        # 18 tokens (vs ~1,800 raw)")
	fmt.Println("  gg chat \"question\"   # AI chat with gg awareness")
	fmt.Println()
	fmt.Println("setup:")
	fmt.Println("  gg init              Configure provider & API key")
	fmt.Println("  gg maaza             Status and setup check")
	fmt.Println("  gg upgrade           Upgrade to Pro ($15/month)")
	fmt.Println("  gg pro               Check Pro subscription status")
	fmt.Println()
	fmt.Println("ai tools:")
	fmt.Println("  gg ask \"...\"         Generate code → PR (Pro)")
	fmt.Println("  gg chat \"...\"        Human-friendly AI chat")
	fmt.Println("  gg a2a [mode]        Agent-to-agent piping (<100 tokens/hop)")
	fmt.Println("  gg edit <file>       AI-assisted file editing")
	fmt.Println("  gg prompts           Manage saved prompts")
	fmt.Println()
	fmt.Println("git:")
	fmt.Println("  gg .                 Current repo → minimal context")
	fmt.Println("  gg user/repo         Any GitHub repo → minimal context")
	fmt.Println("  gg snap              Project state in one line (<20 tokens)")
	fmt.Println("  gg diff [--staged]   Compressed diff (<50 tokens)")
	fmt.Println("  gg log [n] [--stat]  Compressed git log (<30 tokens)")
	fmt.Println("  gg context           Repo profile (<100 tokens)")
	fmt.Println("  gg cat <u/r/path>    Fetch file from GitHub")
	fmt.Println("  gg tree <u/r> [path] Browse repo file tree")
	fmt.Println("  gg pr <number>       View/manage specific PR")
	fmt.Println("  gg approve           Merge PR created by gg ask")
	fmt.Println("  gg test [args]       Compressed test output (<30 tokens)")
	fmt.Println("  gg push [remote]     Compressed push output (~8 tokens)")
	fmt.Println("  gg pull [remote]     Compressed pull output (~10 tokens)")
	fmt.Println("  gg reflog [n]        Compressed reflog (~15 tokens)")
	fmt.Println("  gg cherry-pick <ref> Cherry-pick with conflict summary")
	fmt.Println("  gg revert <ref>      Revert with compressed output")
	fmt.Println("  gg license           Detect and summarize project license")
	fmt.Println("  gg fetch [remote]    Compressed fetch output (~8 tokens)")
	fmt.Println("  gg add <files>       Compressed staging summary (~6 tokens)")
	fmt.Println("  gg commit <msg> [-a] Compressed commit output (~8 tokens)")
	fmt.Println("  gg checkout <ref>    Compressed checkout/switch (~5 tokens)")
	fmt.Println("  gg show [ref]        Compressed show output (~15 tokens)")
	fmt.Println("  gg clean [-f]        Untracked file summary (~15 tokens)")
	fmt.Println("  gg stash [sub]       Compressed stash operations (~6 tokens)")
	fmt.Println("  gg blame <file> [L]  Compressed blame by author (~15 tokens)")
	fmt.Println("  gg tag [action]      Tag list/create/delete (~15 tokens)")
	fmt.Println("  gg merge <branch>    Compressed merge with conflicts (~10 tokens)")
	fmt.Println("  gg worktree [sub]    Worktree list/add/remove (~15 tokens)")
	fmt.Println("  gg reset [mode] [ref] Compressed reset output (~6 tokens)")
	fmt.Println("  gg status            Grouped file status (~4 tokens clean)")
	fmt.Println("  gg branch [sub]      Branch list/create/delete (~15 tokens)")
	fmt.Println("  gg clone <url>       Compressed clone output (~8 tokens)")
	fmt.Println("  gg remote [sub]      Remote list/add/remove (~10 tokens)")
	fmt.Println("  gg run <cmd>         Run command in sandbox")
	fmt.Println()
	fmt.Println("packages:")
	fmt.Println("  gg npm <pkg>         npm → ~18 tokens (vs ~1,800 raw)")
	fmt.Println("  gg pip <pkg>         PyPI → ~18 tokens (vs ~1,200 raw)")
	fmt.Println("  gg go <module>       Go module lookup")
	fmt.Println("  gg crate <name>      Crates.io lookup")
	fmt.Println("  gg brew [-i] <f>     Homebrew → ~22 tokens (vs ~800 raw)")
	fmt.Println("  gg docker <image>    Docker Hub image info")
	fmt.Println("  gg pub <pkg>         pub.dev (Dart/Flutter) lookup")
	fmt.Println("  gg maven <g:a>       Maven Central lookup")
	fmt.Println("  gg nuget <pkg>       NuGet (.NET) lookup")
	fmt.Println("  gg hex <pkg>         Hex.pm (Elixir/Erlang) lookup")
	fmt.Println("  gg gem <pkg>         RubyGems lookup (~18 tokens)")
	fmt.Println("  gg composer <v/p>    Packagist (PHP) lookup (~18 tokens)")
	fmt.Println("  gg search <query>    Cross-registry search (<50 tokens)")
	fmt.Println("  gg deps [file]       Dependency summary (<80 tokens)")
	fmt.Println("  gg chain <tools>     Chain multiple lookups")
	fmt.Println("  gg cool <toolbelt>   Curated toolbelts (webdev, media, sec, data)")
	fmt.Println("  gg cache status      Show cache size")
	fmt.Println()
	fmt.Println("other:")
	fmt.Println("  gg env               Environment summary (~20 tokens)")
	fmt.Println("  gg loc [path]        Lines of code by language (~20 tokens)")
	fmt.Println("  gg todo [path]       Scan TODOs/FIXMEs (~25 tokens)")
	fmt.Println("  gg time [--iso|utc]  Time HUD for agents")
	fmt.Println("  gg mcp               Start MCP tool server (for Claude Desktop, Cursor, etc)")
	fmt.Println("  gg stats             Usage statistics")
	fmt.Println("  gg version           Show version")
	fmt.Println("  gg help              Show this help")
	fmt.Println()
	fmt.Println("note: Local models (Ollama) may vary in accuracy. Use cloud APIs for best results.")
	fmt.Println()
	fmt.Println("install: curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh")
	fmt.Println("docs:    github.com/cyclecore-dev/gg")
}

