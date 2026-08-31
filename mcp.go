package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpTool struct {
	info   mcpToolInfo
	toArgs func(map[string]interface{}) []string
}

func mcpSchema(props map[string]interface{}, required ...string) map[string]interface{} {
	s := map[string]interface{}{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func mcpStr(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": desc}
}

func mcpBool(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "boolean", "description": desc}
}

func mcpInt(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": desc}
}

func mcpNoParams() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}

func mcpParamStr(p map[string]interface{}, key string) string {
	if v, ok := p[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mcpParamBool(p map[string]interface{}, key string) bool {
	if v, ok := p[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func mcpParamFloat(p map[string]interface{}, key string) float64 {
	if v, ok := p[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

var mcpTools = []mcpTool{
	// --- Git ---
	{
		info: mcpToolInfo{Name: "gg_snap", Description: "One-line project state: branch, hash, commit message, status, ahead/behind (<20 tokens)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"snap"} },
	},
	{
		info: mcpToolInfo{Name: "gg_diff", Description: "Compressed git diff showing files changed with insertions/deletions (<50 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"mode": mcpStr("Diff mode: 'staged' for staged changes, 'all' for all changes, or a git ref to diff against"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if mode := mcpParamStr(p, "mode"); mode != "" {
				if mode == "staged" {
					return []string{"diff", "--staged"}
				}
				if mode == "all" {
					return []string{"diff", "--all"}
				}
				return []string{"diff", mode}
			}
			return []string{"diff"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_log", Description: "Compressed git log showing recent commits (<30 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"count":  mcpInt("Number of commits to show (default: 10)"),
			"stat":   mcpBool("Include file change counts per commit"),
			"author": mcpStr("Filter by author name"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"log"}
			if n := mcpParamFloat(p, "count"); n > 0 {
				args = append(args, fmt.Sprintf("%d", int(n)))
			}
			if mcpParamBool(p, "stat") {
				args = append(args, "--stat")
			}
			if a := mcpParamStr(p, "author"); a != "" {
				args = append(args, "--author="+a)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_show", Description: "Compressed commit detail: author, message, files changed (<25 tokens vs ~800 raw)", InputSchema: mcpSchema(map[string]interface{}{
			"ref": mcpStr("Git ref to show (commit hash, branch, tag, HEAD~1, etc). Default: HEAD"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if ref := mcpParamStr(p, "ref"); ref != "" {
				return []string{"show", ref}
			}
			return []string{"show"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_blame", Description: "Compressed git blame showing commit attribution per line range (<30 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"file":  mcpStr("File to blame"),
			"lines": mcpStr("Line range (e.g. '10', '10-20')"),
		}, "file")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"blame", mcpParamStr(p, "file")}
			if lines := mcpParamStr(p, "lines"); lines != "" {
				args = append(args, lines)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_branch", Description: "Branch listing with tracking info and ahead/behind counts (<20 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action": mcpStr("Action: 'list' (default), 'create', or 'delete'"),
			"name":   mcpStr("Branch name (for create/delete)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			name := mcpParamStr(p, "name")
			if action == "" || action == "list" {
				return []string{"branch"}
			}
			if name != "" {
				return []string{"branch", action, name}
			}
			return []string{"branch", action}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_stash", Description: "Compressed stash operations (<15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action":  mcpStr("Action: 'list', 'pop', 'show', 'drop', or a message string to save"),
			"message": mcpStr("Stash message (for save action)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			msg := mcpParamStr(p, "message")
			if msg != "" {
				return []string{"stash", msg}
			}
			if action != "" {
				return []string{"stash", action}
			}
			return []string{"stash"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_tag", Description: "Tag listing and management (<15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action":  mcpStr("Action: 'list' (default), 'create', or 'delete'"),
			"name":    mcpStr("Tag name"),
			"message": mcpStr("Tag message (creates annotated tag)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			name := mcpParamStr(p, "name")
			msg := mcpParamStr(p, "message")
			if action == "delete" && name != "" {
				return []string{"tag", "delete", name}
			}
			if name != "" {
				if msg != "" {
					return []string{"tag", name, msg}
				}
				return []string{"tag", name}
			}
			return []string{"tag"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_remote", Description: "Remote listing with URLs and tracking status (<10 tokens)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"remote"} },
	},
	{
		info: mcpToolInfo{Name: "gg_clean", Description: "Show untracked files with sizes; use force to delete (<15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"force": mcpBool("Actually delete untracked files (default: dry run only)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if mcpParamBool(p, "force") {
				return []string{"clean", "--force"}
			}
			return []string{"clean"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_reflog", Description: "Compressed git reflog showing recent HEAD movements (~15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"count": mcpInt("Number of entries to show (default: 10)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if n := mcpParamFloat(p, "count"); n > 0 {
				return []string{"reflog", fmt.Sprintf("%d", int(n))}
			}
			return []string{"reflog"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_push", Description: "Push with compressed output (~8 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"remote": mcpStr("Remote name (default: origin)"),
			"branch": mcpStr("Branch name (default: current)"),
			"force":  mcpBool("Force push"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"push"}
			if mcpParamBool(p, "force") {
				args = append(args, "--force")
			}
			if r := mcpParamStr(p, "remote"); r != "" {
				args = append(args, r)
			}
			if b := mcpParamStr(p, "branch"); b != "" {
				args = append(args, b)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_pull", Description: "Pull with compressed output showing commits, files, conflicts (~10 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"remote": mcpStr("Remote name (default: origin)"),
			"branch": mcpStr("Branch name (default: current)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"pull"}
			if r := mcpParamStr(p, "remote"); r != "" {
				args = append(args, r)
			}
			if b := mcpParamStr(p, "branch"); b != "" {
				args = append(args, b)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_cherry_pick", Description: "Cherry-pick with conflict summary (~10 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"ref": mcpStr("Commit hash, branch, or tag to cherry-pick"),
		}, "ref")},
		toArgs: func(p map[string]interface{}) []string { return []string{"cherry-pick", mcpParamStr(p, "ref")} },
	},
	{
		info: mcpToolInfo{Name: "gg_revert", Description: "Revert commit with compressed output (~10 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"ref": mcpStr("Commit hash to revert"),
		}, "ref")},
		toArgs: func(p map[string]interface{}) []string { return []string{"revert", mcpParamStr(p, "ref")} },
	},
	{
		info: mcpToolInfo{Name: "gg_license", Description: "Detect and summarize project license (~8 tokens)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"license"} },
	},
	{
		info: mcpToolInfo{Name: "gg_fetch", Description: "Fetch with compressed output (~8 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"remote": mcpStr("Remote name (default: origin)"),
			"all":    mcpBool("Fetch from all remotes"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if mcpParamBool(p, "all") {
				return []string{"fetch", "--all"}
			}
			if r := mcpParamStr(p, "remote"); r != "" {
				return []string{"fetch", r}
			}
			return []string{"fetch"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_add", Description: "Stage files with compressed summary (~6 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"files": mcpStr("Files to stage (space-separated, or '.' for all, '-A' for everything)"),
		}, "files")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"add"}
			if files := mcpParamStr(p, "files"); files != "" {
				args = append(args, strings.Fields(files)...)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_commit", Description: "Commit with compressed output (~8 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"message": mcpStr("Commit message"),
			"all":     mcpBool("Stage all modified tracked files before committing (-a)"),
		}, "message")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"commit", mcpParamStr(p, "message")}
			if mcpParamBool(p, "all") {
				args = append(args, "-a")
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_checkout", Description: "Checkout/switch with compressed output (~5 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"target": mcpStr("Branch, tag, commit, or file to checkout"),
			"create": mcpBool("Create new branch (-b)"),
		}, "target")},
		toArgs: func(p map[string]interface{}) []string {
			if mcpParamBool(p, "create") {
				return []string{"checkout", "-b", mcpParamStr(p, "target")}
			}
			return []string{"checkout", mcpParamStr(p, "target")}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_env", Description: "Environment summary: OS, shell, installed tools (~20 tokens)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"env"} },
	},
	{
		info: mcpToolInfo{Name: "gg_loc", Description: "Lines of code by language (~20 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"path": mcpStr("Directory to scan (default: current directory)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if path := mcpParamStr(p, "path"); path != "" {
				return []string{"loc", path}
			}
			return []string{"loc"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_todo", Description: "Scan TODO/FIXME/HACK/XXX markers (~25 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"path": mcpStr("Directory to scan (default: current directory)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if path := mcpParamStr(p, "path"); path != "" {
				return []string{"todo", path}
			}
			return []string{"todo"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_context", Description: "Full repo profile: language, framework, deps, recent activity (<100 tokens)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"context"} },
	},
	{
		info: mcpToolInfo{Name: "gg_test", Description: "Auto-detect test runner and show compressed results (<30 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"args": mcpStr("Extra arguments passed through to the test runner (e.g. '-v', '-run TestAuth')"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"test"}
			if extra := mcpParamStr(p, "args"); extra != "" {
				args = append(args, strings.Fields(extra)...)
			}
			return args
		},
	},

	// --- Packages ---
	{
		info: mcpToolInfo{Name: "gg_npm", Description: "npm package info compressed to ~18 tokens (vs ~1,800 raw)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("npm package name"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"npm", mcpParamStr(p, "package")} },
	},
	{
		info: mcpToolInfo{Name: "gg_pip", Description: "PyPI package info compressed to ~18 tokens (vs ~1,200 raw)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("PyPI package name"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"pip", mcpParamStr(p, "package")} },
	},
	{
		info: mcpToolInfo{Name: "gg_go", Description: "Go module info compressed to ~18 tokens", InputSchema: mcpSchema(map[string]interface{}{
			"module": mcpStr("Go module path (e.g. 'github.com/gorilla/mux')"),
		}, "module")},
		toArgs: func(p map[string]interface{}) []string { return []string{"go", mcpParamStr(p, "module")} },
	},
	{
		info: mcpToolInfo{Name: "gg_crate", Description: "Crates.io crate info compressed to ~18 tokens", InputSchema: mcpSchema(map[string]interface{}{
			"name": mcpStr("Crate name"),
		}, "name")},
		toArgs: func(p map[string]interface{}) []string { return []string{"crate", mcpParamStr(p, "name")} },
	},
	{
		info: mcpToolInfo{Name: "gg_brew", Description: "Homebrew formula info compressed to ~22 tokens (vs ~800 raw)", InputSchema: mcpSchema(map[string]interface{}{
			"formula": mcpStr("Homebrew formula name"),
			"install": mcpBool("Install the formula if not present"),
		}, "formula")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"brew"}
			if mcpParamBool(p, "install") {
				args = append(args, "-i")
			}
			args = append(args, mcpParamStr(p, "formula"))
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_search", Description: "Cross-registry package search across npm, PyPI, crates.io, and Go proxy (<50 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"query": mcpStr("Search query"),
		}, "query")},
		toArgs: func(p map[string]interface{}) []string { return []string{"search", mcpParamStr(p, "query")} },
	},
	{
		info: mcpToolInfo{Name: "gg_deps", Description: "Dependency summary from project manifest (<80 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"file": mcpStr("Manifest file to parse (auto-detected if omitted)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if f := mcpParamStr(p, "file"); f != "" {
				return []string{"deps", f}
			}
			return []string{"deps"}
		},
	},

	// --- GitHub ---
	{
		info: mcpToolInfo{Name: "gg_cat", Description: "Fetch and display a file from any GitHub repository", InputSchema: mcpSchema(map[string]interface{}{
			"path": mcpStr("File path as user/repo/path/to/file"),
		}, "path")},
		toArgs: func(p map[string]interface{}) []string { return []string{"cat", mcpParamStr(p, "path")} },
	},
	{
		info: mcpToolInfo{Name: "gg_tree", Description: "Browse GitHub repo file tree with sizes (<40 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"repo": mcpStr("Repository as user/repo"),
			"path": mcpStr("Subdirectory path (optional)"),
		}, "repo")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"tree", mcpParamStr(p, "repo")}
			if path := mcpParamStr(p, "path"); path != "" {
				args = append(args, path)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_repo", Description: "GitHub repo summary compressed to ~18 tokens", InputSchema: mcpSchema(map[string]interface{}{
			"repo": mcpStr("Repository as user/repo"),
		}, "repo")},
		toArgs: func(p map[string]interface{}) []string { return []string{mcpParamStr(p, "repo")} },
	},
	{
		info: mcpToolInfo{Name: "gg_current_repo", Description: "Current repo summary compressed to ~12 tokens", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"."} },
	},
	{
		info: mcpToolInfo{Name: "gg_pr", Description: "View PR details: diff summary, comments, status", InputSchema: mcpSchema(map[string]interface{}{
			"number": mcpStr("PR number"),
		}, "number")},
		toArgs: func(p map[string]interface{}) []string { return []string{"pr", mcpParamStr(p, "number")} },
	},

	{
		info: mcpToolInfo{Name: "gg_docker", Description: "Docker Hub image info with pull counts and tags (~20 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"image": mcpStr("Docker image name (e.g. 'nginx' or 'grafana/grafana')"),
		}, "image")},
		toArgs: func(p map[string]interface{}) []string { return []string{"docker", mcpParamStr(p, "image")} },
	},
	{
		info: mcpToolInfo{Name: "gg_pub", Description: "pub.dev (Dart/Flutter) package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("pub.dev package name"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"pub", mcpParamStr(p, "package")} },
	},
	{
		info: mcpToolInfo{Name: "gg_nuget", Description: "NuGet (.NET) package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("NuGet package name (case-insensitive)"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"nuget", mcpParamStr(p, "package")} },
	},
	{
		info: mcpToolInfo{Name: "gg_maven", Description: "Maven Central package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"query": mcpStr("Group:artifact (e.g. 'com.google.guava:guava') or search query"),
		}, "query")},
		toArgs: func(p map[string]interface{}) []string { return []string{"maven", mcpParamStr(p, "query")} },
	},
	{
		info: mcpToolInfo{Name: "gg_hex", Description: "Hex.pm (Elixir/Erlang) package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("Hex package name"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"hex", mcpParamStr(p, "package")} },
	},

	{
		info: mcpToolInfo{Name: "gg_status", Description: "Grouped file status by state (~4 tokens when clean)", InputSchema: mcpNoParams()},
		toArgs: func(p map[string]interface{}) []string { return []string{"status"} },
	},
	{
		info: mcpToolInfo{Name: "gg_gem", Description: "RubyGems package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("RubyGems package name"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"gem", mcpParamStr(p, "package")} },
	},
	{
		info: mcpToolInfo{Name: "gg_merge", Description: "Compressed merge with conflict summary (~10 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"branch": mcpStr("Branch to merge"),
			"no_ff":  mcpBool("Create a merge commit even for fast-forward"),
			"squash": mcpBool("Squash merge (stage changes without committing)"),
		}, "branch")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"merge"}
			if mcpParamBool(p, "no_ff") {
				args = append(args, "--no-ff")
			}
			if mcpParamBool(p, "squash") {
				args = append(args, "--squash")
			}
			args = append(args, mcpParamStr(p, "branch"))
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_worktree", Description: "Worktree list/add/remove (~15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action": mcpStr("Action: 'list' (default), 'add', or 'remove'"),
			"path":   mcpStr("Path for add/remove"),
			"branch": mcpStr("Branch for add (optional)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			path := mcpParamStr(p, "path")
			branch := mcpParamStr(p, "branch")
			if action == "add" && path != "" {
				if branch != "" {
					return []string{"worktree", "add", path, branch}
				}
				return []string{"worktree", "add", path}
			}
			if action == "remove" && path != "" {
				return []string{"worktree", "remove", path}
			}
			return []string{"worktree"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_reset", Description: "Compressed reset output (~6 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"mode": mcpStr("Reset mode: '--soft', '--hard', or '--mixed' (default: --mixed)"),
			"ref":  mcpStr("Commit reference (default: HEAD)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"reset"}
			if mode := mcpParamStr(p, "mode"); mode != "" {
				args = append(args, mode)
			}
			if ref := mcpParamStr(p, "ref"); ref != "" {
				args = append(args, ref)
			}
			return args
		},
	},

	{
		info: mcpToolInfo{Name: "gg_branch", Description: "Branch list/create/delete (~15 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action": mcpStr("Action: 'list' (default), 'create', or 'delete'"),
			"name":   mcpStr("Branch name for create/delete"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			name := mcpParamStr(p, "name")
			if action == "create" && name != "" {
				return []string{"branch", "create", name}
			}
			if action == "delete" && name != "" {
				return []string{"branch", "delete", name}
			}
			return []string{"branch"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_clone", Description: "Compressed clone output (~8 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"url":    mcpStr("Git URL or GitHub user/repo shorthand"),
			"target": mcpStr("Target directory (optional)"),
		}, "url")},
		toArgs: func(p map[string]interface{}) []string {
			args := []string{"clone", mcpParamStr(p, "url")}
			if t := mcpParamStr(p, "target"); t != "" {
				args = append(args, t)
			}
			return args
		},
	},
	{
		info: mcpToolInfo{Name: "gg_remote", Description: "Remote list/add/remove with tracking status (~10 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"action": mcpStr("Action: 'list' (default), 'add', or 'rm'"),
			"name":   mcpStr("Remote name for add/rm"),
			"url":    mcpStr("Remote URL for add"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			action := mcpParamStr(p, "action")
			name := mcpParamStr(p, "name")
			url := mcpParamStr(p, "url")
			if action == "add" && name != "" && url != "" {
				return []string{"remote", "add", name, url}
			}
			if action == "rm" && name != "" {
				return []string{"remote", "rm", name}
			}
			return []string{"remote"}
		},
	},
	{
		info: mcpToolInfo{Name: "gg_composer", Description: "Packagist (PHP) package lookup (~18 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"package": mcpStr("Packagist package name (vendor/package)"),
		}, "package")},
		toArgs: func(p map[string]interface{}) []string { return []string{"composer", mcpParamStr(p, "package")} },
	},

	// --- Other ---
	{
		info: mcpToolInfo{Name: "gg_time", Description: "Current date/time optimized for AI agents (<8 tokens)", InputSchema: mcpSchema(map[string]interface{}{
			"format": mcpStr("Format: 'iso', 'utc', 'unix', 'week' (default: human-readable)"),
		})},
		toArgs: func(p map[string]interface{}) []string {
			if f := mcpParamStr(p, "format"); f != "" {
				return []string{"time", "--" + f}
			}
			return []string{"time"}
		},
	},
}

func handleMCP() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Content-Length") {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		if req.ID == nil {
			continue
		}

		var resp rpcResponse
		resp.JSONRPC = "2.0"
		resp.ID = req.ID

		switch req.Method {
		case "initialize":
			resp.Result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":      map[string]interface{}{"name": "gg", "version": version},
			}
		case "ping":
			resp.Result = map[string]interface{}{}
		case "tools/list":
			tools := make([]mcpToolInfo, len(mcpTools))
			for i, t := range mcpTools {
				tools[i] = t.info
			}
			resp.Result = map[string]interface{}{"tools": tools}
		case "resources/list":
			resp.Result = map[string]interface{}{"resources": []interface{}{}}
		case "prompts/list":
			resp.Result = map[string]interface{}{"prompts": []interface{}{}}
		case "tools/call":
			resp.Result = mcpCallTool(req.Params)
		default:
			resp.Error = &rpcError{Code: -32601, Message: "Method not found: " + req.Method}
		}

		out, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stderr, "[gg-mcp] %s → %s\n", req.Method, req.ID)
		os.Stdout.Write(out)
		os.Stdout.Write([]byte("\n"))
	}
}

func mcpCallTool(raw json.RawMessage) interface{} {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return map[string]interface{}{
			"content": []mcpContent{{Type: "text", Text: "Invalid parameters"}},
			"isError": true,
		}
	}

	for _, tool := range mcpTools {
		if tool.info.Name == params.Name {
			if params.Arguments == nil {
				params.Arguments = map[string]interface{}{}
			}
			args := tool.toArgs(params.Arguments)
			output := mcpExecGG(args)
			return map[string]interface{}{
				"content": []mcpContent{{Type: "text", Text: output}},
			}
		}
	}

	return map[string]interface{}{
		"content": []mcpContent{{Type: "text", Text: "Unknown tool: " + params.Name}},
		"isError": true,
	}
}

func mcpExecGG(args []string) string {
	exe, err := os.Executable()
	if err != nil {
		return "Error: cannot locate gg binary"
	}
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), "GG_MCP_CHILD=1")
	output, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(output))
	if err != nil && result == "" {
		return "Error: " + err.Error()
	}
	return result
}
