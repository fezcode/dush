# Instructions

- Read the README.md to get the structure first.
- No external libs will be used, when we need something we will implement. The only exception to this is `github.com/fezcode/go-piml`
- Every directory has README.md file that describes its purpose. If you still do not understand it, then read below.
- Write understandable comments especially about niche things. 
- Config dir is `~/.dush`
- use build scripts to build the app.

## Directory Structure
- dush/: Project root.
- dush.exe: Compiled executable.
- go.mod: Go module definition and dependencies.
- go.sum: Dependency checksums.
- README.md: Project overview, setup, and structure guide.
- .idea/: Ignore this folder it is for IntelliJ IDEA configuration files (contains .gitignore, .iml, modules.xml, workspace.xml).
- cmd/: Main packages for executables.
    - cmd/dush/: dush executable's main package.
        - cmd/dush/main.go: Application entry point.
        - cmd/dush/commands/: CLI command definitions (e.g., root command in root.go).
- internal/: Private application code.
    - internal/builtins/: Shell built-in command implementations.
    - internal/evaluator/: Command interpretation and execution logic.
    - internal/parser/: Command line input parsing.
    - internal/repl/: Read-Eval-Print Loop logic.
    - internal/utils/: General utility functions.
- pkg/: (Currently empty) Intended for reusable public libraries.
- scripts/: Automation scripts.
- test/: Project test files.
