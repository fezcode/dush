# Feature Items

Features to make dush a better shell. Not chasing bash compatibility — building something better.

## Legend

- `[x]` Done
- `[-]` Already doable (workaround exists, see notes)
- `[ ]` Not yet implemented
- `ROADMAP` = was on the language spec roadmap

---

## Data Types & Literals

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Array literals `[1, 2, 3]` | `[x]` | Full support with push/pop/slice/map/filter/reverse/first/last methods. |
| 2 | Map/dict type `{"key": "value"}` | `[x]` | Full support with keys/values/has/delete/merge methods. Insertion-order preserved. |
| 3 | Index expressions `@arr[0]`, `@map["key"]` | `[x]` | Works on arrays, maps, and strings. Negative indices wrap. Assignment supported. |
| 4 | Range literals `1..10`, `1..=10` | `[x]` | `1..5` exclusive, `1..=5` inclusive. Returns array, works in loops. See `tests/ranges.dush`. |
| 5 | Null literal | `[-]` | Null exists internally but no `null` keyword. Rarely needed — dush doesn't have nullable vars. |

## Shell Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 6 | Globbing in commands `ls *.go` | `[x]` | Already works — `evalCommandExpression` calls `filepath.Glob` on args containing `*` or `?`. |
| 7 | Tilde expansion `~/file` | `[x]` | Already works — `evalCommandExpression` expands `~` to home dir. |
| 8 | Stderr redirect `2>`, `2>>`, `&>` | `[x]` | `2>file`, `2>>file`, `&>file` (both stdout+stderr). |
| 9 | Here-docs `<<EOF ... EOF` | `[x]` | Multi-line string input to commands. See `tests/heredocs.dush`. |
| 10 | Here-strings `<<<` | `[x]` | Feed a string directly as stdin: `sort <<< "hello"`. |
| 11 | Subshells `(cmd1; cmd2)` | `[ ]` | Run commands in an isolated environment. Useful for temp cd, var changes. |
| 12 | Process substitution `<(cmd)` | `[ ]` | Treat command output as a file path. Niche but powerful for diff, paste, etc. |
| 13 | Brace expansion `file.{txt,md}` | `[ ]` | Generates permutations. Useful for bulk file ops. |
| 14 | `exec` builtin | `[ ]` | Replace current shell process / redirect file descriptors. |
| 15 | `PROMPT_COMMAND` / prompt hook | `[ ]` | Run code before every prompt. Useful for dynamic prompts (git branch, etc). |

## Error Handling

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 16 | Try/catch blocks | `[ ]` ROADMAP | `try { risky() } catch (@err) { echo @err }`. Way better than bash `trap ERR`. |
| 17 | Result type / error propagation | `[ ]` | Optional: procs return `ok(val)` or `err(msg)`. Overkill for a shell? |

## Control Flow

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 18 | `else if` chains | `[-]` | Works now: `if (...) { } else { if (...) { } }` — but no `else if` shorthand. Could be nice. |
| 19 | `break` / `continue` in loops | `[x]` | Both keywords work in `loop` blocks. |
| 20 | Loop `else` clause | `[ ]` | `loop (...) { } else { }` — runs else if loop body never executed. Python-style. |

## String Features

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 21 | Regex match with captures | `[x]` | `.match()` returns captures, `.matches()` returns bool, `.match_all()` returns all matches, `.replace_regex()` replaces by pattern. RE2 syntax. See `tests/regex.dush`. |
| 22 | Multi-line strings | `[-]` | Double-quoted strings already span lines. Raw strings too. Works. |
| 23 | String repeat | `[ ]` | `@s.repeat(3)` or `"ab" * 3`. Small utility. |
| 24 | String indexing | `[x]` | `@s[0]` works via index expressions (#3). Negative indices wrap. |

## Built-in Utilities

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 25 | Built-in JSON parse/stringify | `[ ]` | `json_parse(@str)` → map, `json_string(@map)` → string. Every modern script needs this. |
| 26 | Built-in HTTP fetch | `[ ]` | `fetch("https://...")` returning response object. Avoid curl flag memorization. |
| 27 | `env` builtin | `[-]` | `pub` exports vars. To list all env vars, user can run external `env` command. Could add `env` builtin to list dush's exported vars. |
| 28 | `test` / file tests `-f`, `-d`, `-z` | `[-]` | `exists(path)` and `is_dir(path)` cover file tests. `-z` (empty string) = `@s.len() == 0`. No dedicated `test` command needed. See `tests/file_tests.dush`. |
| 29 | Arithmetic expansion | `[-]` | `echo (@x + 1)` already works via grouped expression args. See `tests/arithmetic.dush`. |
| 30 | `getopts` / flag parsing | `[ ]` | Parse flags inside scripts. Could be a `flags()` builtin or a proc pattern. |

## Scripting Power

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 31 | `strict on` / `strict on { ... }` | `[x]` | Abort on first non-zero exit. Bare form = scope-wide, block form = scoped (block aborts but the outer script continues). |
| 32 | `trace on` / `trace on { ... }` | `[x]` | Print each command (with resolved args) to stderr prefixed `+ ` before running it. |
| 32a | `pipefail on` | `[x]` | Pipeline status = first non-zero stage instead of last. |
| 33 | `eval` (dynamic code execution) | `[ ]` | `eval("@x + 1")`. Dangerous but sometimes needed. |
| 34 | Programmable tab completion API | `[ ]` | Let users define completions for their own procs. |
| 35 | `mapfile` / read lines into array | `[ ]` | `@lines = lines("file.txt")` or `save(cat file) | split("\n")`. Needs array literals first. |

## Already Working (for reference)

These are sometimes listed as "missing" but dush already handles them:

| Feature | How it works in dush |
|---------|---------------------|
| Globbing `*.go` | Automatic in command args via `filepath.Glob` |
| Tilde `~` expansion | Automatic in command args |
| Pipes `\|` | `cmd1 \| cmd2` |
| Redirects `>`, `>>`, `<` | `cmd > file`, `cmd >> file`, `cmd < file` |
| Background `&` | `cmd &`, `jobs`, `fg`, `killjob` |
| Signal trapping | `trap 'cmd' INT` |
| Local scope | `let @x = 10` |
| Readonly vars | `const @PI = 3.14` |
| Export vars | `pub @KEY = "val"` |
| Closures | Procs capture enclosing env |
| Output capture | `save(cmd)` |
| Match/case | Pattern matching with wildcards |

---

## Priority Order

**Phase 1 — Core data types (do first, unblocks everything):**
1. Array literals (#1)
2. Index expressions (#3)
3. Map type (#2)

**Phase 2 — Shell power:**
4. Stderr redirects (#8)
5. Here-docs (#9) + here-strings (#10)
6. Subshells (#11)
7. Break/continue (#19)

**Phase 3 — Modern shell extras:**
8. Try/catch (#16)
9. Regex match (#21)
10. JSON parse/stringify (#25)
11. HTTP fetch (#26)
12. Else-if shorthand (#18)

**Phase 4 — Polish:**
13. Brace expansion (#13)
14. Process substitution (#12)
15. Debug tracing (#32)
16. Programmable completion (#34)
17. Fail-fast mode (#31)
