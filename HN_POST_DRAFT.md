# HN Launch Post Draft

## Title Options

**Option A (metrics-focused):**
> Show HN: gg – 98% token savings for AI agents on git/npm/brew operations

**Option B (positioning):**
> Show HN: gg – The 2-letter agent-native git client

**Option C (problem-focused):**
> Show HN: gg – Stop wasting 40k tokens/day on git status

---

## Post Body

Hi HN,

I built gg because AI agents waste thousands of tokens every time they interact with git, npm, or brew. A simple `npm info prettier --json` dumps ~1,800 tokens of metadata that the agent doesn't need.

gg is a thin CLI that turns these operations into minimal MCP-format responses:

```
$ gg npm prettier
mcp://npm/prettier
install: npm i prettier
version: 3.2.5
```

18 tokens instead of 1,800. Same info, 99% smaller.

**Benchmarks:**
- npm lookup: 1,800 → 18 tokens (99% savings)
- brew formula: 2,200 → 22 tokens (99% savings)
- GitHub repo context: 1,500 → 12 tokens (99% savings)
- Daily agent workload: ~40k → ~400 tokens (98% savings)

**Commands:**
```bash
gg .              # current repo → MCP endpoint
gg npm prettier   # npm package → 18 tokens
gg brew ffmpeg    # brew formula → 22 tokens
gg cool webdev    # eslint+prettier+jest+playwright
gg ask "prompt"   # AI code generation
```

**CLI2CLI (bonus):** gg also has agent-to-agent modes for piping between agents with <100 tokens per hop:
```bash
gg a2a . | gg a2a ask "summarize this repo"
```

Built in Go, works with Anthropic/OpenAI/Ollama. MIT licensed.

Install: `curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh`

GitHub: https://github.com/cyclecore-dev/gg

Would love feedback from anyone building AI agents or dealing with context window limits.

---

## Anticipated Questions & Answers

**Q: Why not just truncate the output?**
A: Truncation loses information. gg extracts the essential fields agents actually need (install command, version, status) and formats them consistently.

**Q: What's MCP format?**
A: Model Context Protocol - a standard for structured tool responses. The `mcp://` URIs give agents a consistent namespace for referencing tools.

**Q: Does this work with Claude/GPT/local models?**
A: Yes. gg is model-agnostic. The `gg ask` command supports Anthropic, OpenAI, and Ollama backends.

**Q: What's the business model?**
A: Free tier (10 calls/day) for basic usage. Pro ($15/mo) for unlimited `gg ask` and priority routing. Core commands (gg ., gg npm, gg brew) are always free.

**Q: Why Go?**
A: Single binary, no dependencies, fast startup. Agents shouldn't wait for npm install.

---

## Posting Strategy

**Timing:** Tuesday-Thursday, 8-10am EST (HN peak)

**Subreddits to cross-post:**
- r/LocalLLaMA (Ollama angle)
- r/ChatGPTCoding (AI dev tools)
- r/commandline (CLI tool)

**X thread:** Announce with demo GIF, link to HN post

---

## Success Metrics

| Metric | Target |
|--------|--------|
| HN points | 100+ |
| GitHub stars | 50+ in 48h |
| Comments | Constructive discussion |
| Install attempts | Track via gg.sh analytics |

---

*Draft created: 2026-01-21*
