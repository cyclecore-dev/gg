# gg — the 2-letter agent-native git client

AI agents waste thousands of tokens on git, npm, and brew calls because raw CLI output is bloated with metadata. gg compresses it: **96–98% token savings**, structured JSON output, and an MCP server with 55 tools.

72 commands. 13 package ecosystems. One binary.

## Install

```bash
# Homebrew (macOS/Linux)
brew install cyclecore-dev/gg/gg

# curl (any platform)
curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh

# npm
npm install -g @ggdotdev/gg

# Arch Linux (AUR)
yay -S gg-bin

# Build from source
git clone https://github.com/cyclecore-dev/gg && cd gg && go build -o gg
```

## Quick Start

```bash
gg init                    # configure provider & API key
gg npm prettier            # npm lookup → ~18 tokens (vs ~1,800 raw)
gg snap                    # repo state in one line (<20 tokens)
gg diff --staged --json    # structured diff for agents
gg why handleAuth          # causal analysis — who, when, why
gg verify --build          # one-line test/lint/build summary
gg mcp                     # start MCP tool server for Claude Desktop / Cursor
```

## Token Savings

| Scenario | Without gg | With gg | Savings |
|----------|-----------|---------|---------|
| npm package lookup | ~1,800 tokens | ~18 tokens | **99%** |
| GitHub file + PR | ~2,400 tokens | ~62 tokens | **97%** |
| Daily agent (20 calls) | 40k tokens | 800 tokens | **98%** |

## Commands

### Git — Read

| Command | Description | ~Tokens |
|---------|-------------|---------|
| `gg .` | Current repo context | ~12 |
| `gg user/repo` | Any GitHub repo context | ~18 |
| `gg snap` | Project state in one line | <20 |
| `gg diff [--staged\|--all]` | Compressed diff | <50 |
| `gg log [n] [--stat]` | Compressed git log | <30 |
| `gg context` | Repo profile | <100 |
| `gg cat <u/r/path>` | Fetch file from GitHub | varies |
| `gg tree <u/r> [path]` | Browse repo file tree | varies |
| `gg show [ref]` | Compressed show output | ~15 |
| `gg reflog [n]` | Compressed reflog | ~15 |
| `gg blame <file> [L]` | Blame compressed by author | ~15 |
| `gg status` | Grouped file status | ~4 |
| `gg license` | Detect and summarize license | ~15 |
| `gg why <query>` | Causal summary (who/when/why) | ~35 |

### Git — Write

| Command | Description | ~Tokens |
|---------|-------------|---------|
| `gg add <files>` | Staging summary | ~6 |
| `gg commit <msg> [-a]` | Commit output | ~8 |
| `gg push [remote]` | Push output | ~8 |
| `gg pull [remote]` | Pull output | ~10 |
| `gg fetch [remote]` | Fetch output | ~8 |
| `gg checkout <ref>` | Checkout/switch | ~5 |
| `gg cherry-pick <ref>` | Cherry-pick with conflict summary | ~10 |
| `gg revert <ref>` | Revert output | ~10 |
| `gg merge <branch>` | Merge with conflict summary | ~10 |
| `gg reset [mode] [ref]` | Reset output | ~6 |
| `gg stash [sub]` | Stash operations | ~6 |
| `gg clean [-f]` | Untracked file summary | ~15 |
| `gg tag [action]` | Tag list/create/delete | ~15 |

### Git — Management

| Command | Description | ~Tokens |
|---------|-------------|---------|
| `gg pr <number>` | View/manage specific PR | ~22 |
| `gg approve` | Merge PR created by `gg ask` | ~8 |
| `gg branch [sub]` | Branch list/create/delete | ~15 |
| `gg clone <url>` | Clone output | ~8 |
| `gg remote [sub]` | Remote list/add/remove | ~10 |
| `gg worktree [sub]` | Worktree list/add/remove | ~15 |

### Package Managers (13 ecosystems)

| Command | Registry | ~Tokens |
|---------|----------|---------|
| `gg npm <pkg>` | npm | ~18 |
| `gg pip <pkg>` | PyPI | ~18 |
| `gg go <module>` | Go modules | ~18 |
| `gg crate <name>` | Crates.io | ~18 |
| `gg brew [-i] <formula>` | Homebrew | ~22 |
| `gg docker <image>` | Docker Hub | ~18 |
| `gg pub <pkg>` | pub.dev (Dart/Flutter) | ~18 |
| `gg maven <group:artifact>` | Maven Central | ~18 |
| `gg nuget <pkg>` | NuGet (.NET) | ~18 |
| `gg hex <pkg>` | Hex.pm (Elixir/Erlang) | ~18 |
| `gg gem <pkg>` | RubyGems | ~18 |
| `gg composer <vendor/pkg>` | Packagist (PHP) | ~18 |
| `gg search <query>` | Cross-registry search | <50 |

### AI Tools

| Command | Description |
|---------|-------------|
| `gg ask "prompt"` | Generate code and open a PR (Pro) |
| `gg chat "question"` | Human-friendly AI chat |
| `gg edit <file>` | AI-assisted file editing |
| `gg prompts` | Manage saved prompts |

### Agent-to-Agent (`gg a2a`)

Minimal, pipeable output for agent-to-agent communication. <100 tokens per hop.

| Command | Description |
|---------|-------------|
| `gg a2a .` | Current repo as minimal endpoint |
| `gg a2a ask "prompt"` | Structured response, no git ops |
| `gg a2a plan "task"` | Numbered plan steps |
| `gg a2a code "task"` | Raw code only, no prose |

```bash
gg a2a . | gg a2a ask "summarize this repo"
gg a2a plan "auth system" | gg a2a code
```

### Analysis & Testing

| Command | Description | ~Tokens |
|---------|-------------|---------|
| `gg test [args]` | Compressed test output | <30 |
| `gg verify [--build]` | One-line test/lint/build | ~10 |
| `gg deps [file]` | Dependency summary | <80 |
| `gg loc [path]` | Lines of code by language | ~20 |
| `gg todo [path]` | Scan TODOs/FIXMEs | ~25 |
| `gg run <cmd>` | Run command in sandbox | varies |
| `gg chain <tools>` | Chain multiple lookups | varies |
| `gg cool <toolbelt>` | Curated toolbelts | ~90 |

### Setup & Account

| Command | Description |
|---------|-------------|
| `gg init` | Configure provider & API key |
| `gg config` | View/edit configuration |
| `gg maaza` | Status and setup check |
| `gg upgrade` | Upgrade to Pro ($15/month) |
| `gg pro` | Check Pro subscription status |

### System

| Command | Description |
|---------|-------------|
| `gg health` | Health check for extensions |
| `gg ensure` | Version handshake for plugins |
| `gg mcp` | Start MCP tool server (55 tools) |
| `gg env` | Environment summary |
| `gg time [--iso\|utc]` | Time HUD for agents |
| `gg cache status` | Cache size |
| `gg stats` | Usage statistics |
| `gg version` | Show version |
| `gg help` | Show help |

## Structured Output

Most commands support `--json` for structured output and `--max-tokens N` for truncation:

```bash
gg snap --json
# {"branch":"main","hash":"a1b2c3d","message":"fix auth","status":"clean","upstream":"↑0 ↓0","clean":true}

gg diff --staged --json
# {"files":[{"path":"auth.go","add":30,"del":8,"type":"M"}],"add":30,"del":8,"count":1}

gg log 5 --json
# {"count":5,"commits":[{"hash":"a1b2c3d","message":"fix auth"}, ...]}

gg diff --max-tokens 20
# 3 files | +47 -12
#   M src/auth.go (+30…
```

## MCP Server

`gg mcp` starts a JSON-RPC 2.0 stdio server with 55 tools. Works with Claude Desktop, Cursor, Windsurf, and any MCP-compatible client.

```json
{
  "mcpServers": {
    "gg": {
      "command": "gg",
      "args": ["mcp"]
    }
  }
}
```

## Toolbelts (`gg cool`)

| Belt | Tools |
|------|-------|
| `webdev` | eslint, prettier, typescript, jest, playwright |
| `media` | ffmpeg, imagemagick, exiftool |
| `sec` | semgrep, snyk, trivy |
| `data` | duckdb, jq, csvtojson |
| `devops` | terraform, kubectl, docker |

## Multi-Provider Support

| Provider | Models |
|----------|--------|
| Anthropic | claude-sonnet-4, claude-opus-4 |
| OpenAI | gpt-4o, gpt-4-turbo |
| Ollama | llama3, codellama, etc. (local) |

Configure via `gg init` or set directly in `~/.gg/config.toml`.

## Free vs Pro

| Feature | Free | Pro ($15/mo) |
|---------|------|--------------|
| All 72 commands | Yes | Yes |
| `gg ask` (AI code + PR) | Rate-limited | Unlimited + priority |
| `gg edit` / `gg prompts` | Basic | Full AI-assisted |
| Token savings | Full (96–98%) | Full + faster responses |

No credit card needed to start.

## License

MIT — [github.com/cyclecore-dev/gg](https://github.com/cyclecore-dev/gg)

---

Independent open-source software sponsored by [CycleCore Technologies](https://cyclecore.ai).
Providers and settings configurable via `~/.gg/config.toml`. No warranty.
