# gg — the 2-letter agent-native git client

## Why gg?

AI agents waste thousands of tokens every time they interact with git, npm, or brew because full schemas and metadata bloat the context. gg fixes that.

It acts as a **deterministic command broker**: a thin, fast bridge that turns complex repo/package operations into tiny MCP calls (11–62 tokens instead of 1,800–2,400).

Result: up to **96–98% token savings** on git/npm/brew tasks — so agents stay fast, cheap, and focused on what matters.

gg is:
- **Independent open-source** (MIT)
- **Model-agnostic** (Anthropic, OpenAI, Ollama)
- **Agent-first UX** (2 letters, chainable, Pro tier for unlimited ask)

Built by the CycleCore team — privacy-first AI infrastructure.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh
```

## Quick Start

```bash
gg init                 # configure provider & API key
gg npm prettier         # npm package → MCP (~18 tokens)
gg brew ffmpeg          # Homebrew → MCP (~22 tokens)
gg cool webdev          # eslint+prettier+jest+playwright
gg edit main.go         # AI-assisted file editing
```

## Commands

### Core (v0.9.6)

| Command | Description | Tokens |
|---------|-------------|--------|
| `gg init` | Configure provider/API key | - |
| `gg maaza` | Status + setup check | - |
| `gg version` | Show version | - |
| `gg stats` | Usage statistics | - |

### Package Manager

| Command | Description | Tokens |
|---------|-------------|--------|
| `gg npm <pkg>` | npm package → MCP | ~18 |
| `gg brew [-i] <formula>` | Homebrew formula (-i auto-installs) | ~22 |
| `gg chain <tools>` | Chain multiple MCPs | variable |
| `gg chain run <name>` | Execute saved chain | variable |
| `gg cool <toolbelt>` | Curated toolbelts | ~90 |
| `gg cache status` | Show cache size | - |
| `gg cache clean` | Prune old entries | - |

### Git Operations

| Command | Description | Tokens |
|---------|-------------|--------|
| `gg .` | Current repo → MCP | ~12 |
| `gg user/repo` | Any GitHub repo → MCP | ~18 |
| `gg pr <number>` | View/manage PR | ~22 |
| `gg run <cmd>` | Sandbox execution | ~15 |

### AI Tools

| Command | Description |
|---------|-------------|
| `gg edit <file>` | AI-assisted file editing |
| `gg ask "prompt"` | Generate code + PR |
| `gg prompts` | Manage saved prompts |
| `gg prompts add <name>` | Save a prompt |
| `gg prompts run <name>` | Execute saved prompt |

### CLI2CLI: Agent-to-Agent Modes (v0.9.6)

Minimal, pipeable CLI modes for agent-to-agent communication. Output is structured, <100 tokens per hop.

| Command | Description | Tokens |
|---------|-------------|--------|
| `gg a2a .` | Current repo → minimal endpoint | ~5 |
| `gg a2a ask "prompt"` | Structured response, no git ops | <100 |
| `gg a2a plan "task"` | Numbered plan steps | <100 |
| `gg a2a code "task"` | Raw code only, no prose | varies |

**Pipe examples:**

```bash
# Agent chain: get repo → ask question
gg a2a . | gg a2a ask "summarize this repo"

# Plan then code
gg a2a plan "auth system" | gg a2a code
```

**Design principles:**
- <100 tokens per output (agents see minimal text)
- YAML frontmatter for structured parsing
- Pipeable stdout (Unix philosophy)
- `--help` for discovery

### Toolbelts (`gg cool`)

| Belt | Tools |
|------|-------|
| `webdev` | eslint, prettier, typescript, jest, playwright |
| `media` | ffmpeg, imagemagick, exiftool |
| `sec` | semgrep, snyk, trivy |
| `data` | duckdb, jq, csvtojson |
| `devops` | terraform, kubectl, docker |

## Multi-Provider Support

gg works with multiple AI providers:

| Provider | API Key Format | Models |
|----------|----------------|--------|
| Anthropic | `sk-ant-*` | claude-sonnet-4, claude-opus-4 |
| OpenAI | `sk-*` | gpt-4o, gpt-4-turbo |
| Ollama | (local) | llama3, codellama, etc. |

Configure via `gg init` or set in `~/.config/gg/config.toml`.

## Token Savings

| Scenario | Without gg | With gg | Savings |
|----------|-----------|---------|---------|
| npm package lookup | ~1,800 tokens | ~18 tokens | **99%** |
| GitHub file + PR | ~2,400 tokens | ~62 tokens | **97%** |
| Daily agent (20 calls) | 40k tokens | 800 tokens | **98%** |

## Examples

```bash
# Configure (first time)
gg init

# Chain npm + brew tools
gg chain npm:prettier npm:eslint brew:jq
gg chain --save webformat npm:prettier npm:eslint
gg chain run webformat

# Auto-install missing Homebrew formula
gg brew -i ffmpeg

# AI-assisted editing
gg edit src/main.go

# Saved prompts
gg prompts add review "Review this code for bugs"
gg prompts run review
```

## Build from Source

```bash
git clone https://github.com/cyclecore-dev/gg
cd gg && go build -o gg
./gg version
```

## Free vs Pro

| Feature | Free | Pro ($15/mo) |
|---------|------|--------------|
| Basic commands | gg ., gg maaza, gg npm, gg brew, etc. | All |
| gg ask (AI code/PR gen) | Rate-limited | Unlimited + priority routing |
| gg edit / gg prompts | Basic | Full AI-assisted editing |
| Token savings | Full (96–98% on git/npm/brew) | Full + faster responses |
| Private repos / advanced | — | Coming soon (SOC2, audit logs) |

Pro unlocks unlimited `gg ask` and priority API routing for heavy agent use.

No credit card needed to start — install and go.

## License

MIT — [github.com/cyclecore-dev/gg](https://github.com/cyclecore-dev/gg)

---

gg is independent open-source software.
Default providers and backends are configurable via `~/.gg/config.toml`.
Built by the CycleCore team — privacy-first AI infrastructure.
No warranty. Use at your own risk.
