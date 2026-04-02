# Dush Development Progress - Summary & Next Steps

This document summarizes the state of the `dush` shell development as of March 16, 2026.

## Completed Features 🚀

### 1. Enhanced Built-in Library
- **String Manipulation**: Added `split`, `replace`, and `to_upper`.
- **FS Utilities**: Added `exists` and `is_dir`.
- **Output Capture**: Implemented `save(command)` to capture stdout into a string variable.

### 2. Standard Shell Features
- **Pipes (`|`)**: Full support for piping stdout from one process/builtin to another.
- **Redirections**:
    - Output Redirection: `>` (overwrite) and `>>` (append).
    - Input Redirection: `<`.
- **Globbing**: Automatic expansion of wildcard patterns (`*`, `?`) in command arguments using `filepath.Glob`.

### 3. Advanced Language Constructs
- **Scoped Environment blocks**: Added `with (KEY="value", ...) { ... }` which injects environment variables only for the duration of the block. Use `env` to verify locally.
- **Improved Scripting**: Loops, procedures, and assignments are fully integrated with the new shell features.

### 4. Build & Standards
- Added `-v/--version` and `-h/--help` flags to the binary.
- Integrated `gobake` build recipe (though the project is not strictly bound to the Atlas HUB standards per user request).
- Fixed script build constraints (`//go:build ignore`) to prevent compilation errors in `scripts/`.

---

## Technical Details

- **Evaluator**: `evaluator.go` now handles `InfixExpression` for shell operations (`|`, `>`, etc.) by detecting if the left side is a command or identifier.
- **Environment**: Added `Stdin io.Reader` and `EnvOverrides map[string]string` to the `Environment` struct to facilitate pipelines and `with` blocks.
- **Parser**: Updated precedence rules to include `PIPE` and `APPEND`. Added `parseWithExpression` as a prefix parsing function.

---

## How to Verify

### Standard Tests
```powershell
go test ./...
```

### Interactive Manual Verification
Run dush with `go run .\cmd\dush\`:

1. **Globbing**: `echo *.go`
2. **Pipes & Redirection**: `cat < README.md | findstr Dush > summary.txt`
3. **With Blocks**: `with (MSG="hello") { echo MSG }`
4. **Output Capture**: `let content = save(ls); print(content)`

---

## Where I Left Off / Pending Work

- [x] Implement Pipes/Redirection logic.
- [x] Implement Globbing expansion.
- [x] Implement `with` scoped blocks.
- [ ] **Cross-Platform Verification**: While logic is OS-agnostic, behavior on Linux vs Windows (like path separators in globbing) should be monitored.
- [ ] **Pipe Concurrency**: The current pipe implementation uses a goroutine for the left side. Investigating if built-ins need special handling when they block on Stdin.
- [ ] **Additional Built-ins**: The `PROC` implementation in `parser.go` has a "TODO", though procedures work via `proc name() {}` statements.

---
**Status**: Core roadmap features implemented. Ready for user feedback or final polishing.
