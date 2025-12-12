# gg — the agent-native package manager

gg turns npm, Homebrew, and GitHub into instant MCPs.
One word = 5 tools chained. 98% token savings.

## Install

```bash
curl -L gg.sh | sh
```

## Quick Start

```bash
gg npm prettier       # npm package → MCP (~18 tokens)
gg brew ffmpeg        # Homebrew → MCP (~22 tokens)
gg cool webdev        # eslint+prettier+jest+playwright
gg chain run mytools  # execute saved chains
```

## Commands

### Package Manager (v0.7)

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
| `gg ask "..."` | Generate code → PR | variable |
| `gg approve` | Merge PR | ~8 |
| `gg pr <number>` | View/manage PR | ~22 |
| `gg run <cmd>` | Sandbox execution | ~15 |
| `gg stats` | Usage statistics | - |

### Toolbelts (`gg cool`)

| Belt | Tools |
|------|-------|
| `webdev` | eslint, prettier, typescript, jest, playwright |
| `media` | ffmpeg, imagemagick, exiftool |
| `sec` | semgrep, snyk, trivy |
| `data` | duckdb, jq, csvtojson |
| `devops` | terraform, kubectl, docker |

## Why gg?

| Scenario | Without gg | With gg | Savings |
|----------|-----------|---------|---------|
| npm package lookup | ~1,800 tokens | ~18 tokens | **99%** |
| GitHub file + PR | ~2,400 tokens | ~62 tokens | **97%** |
| Daily agent (20 calls) | 40k tokens | 800 tokens | **98%** |

## Examples

```bash
# Chain npm + brew tools
gg chain npm:prettier npm:eslint brew:jq
gg chain --save webformat npm:prettier npm:eslint
gg chain run webformat

# Auto-install missing Homebrew formula
gg brew -i ffmpeg

# View cache
gg cache status

# Use a curated toolbelt
gg cool webdev
```

## Build from Source

```bash
git clone https://github.com/ggdotdev/gg
cd gg && go build -o gg
./gg version
```

## License

MIT — [github.com/ggdotdev/gg](https://github.com/ggdotdev/gg)
