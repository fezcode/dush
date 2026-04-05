# Dush

A modern shell written in Go with a scripting language that actually makes sense.

Dush replaces the cryptic `${}` syntax of bash with a clean `@` sigil system. Variables are explicit, methods are chainable, and commands just work.

## Quick Look

```bash
@name = "world"
echo "hello @name"             # hello world
echo @name.upper()             # WORLD

const @PI = 3.14               # immutable
pub @KEY = "abc"               # exported to child processes

match (@status) {
    case 0 { echo "ok" }
    case _ { echo "fail" }
}

proc greet(@who) {
    echo "hey @who!"
}
greet("dush")

sleep 30 &                     # background job
jobs                           # list running jobs
```

## Features

- **`@` variable system** -- unambiguous access everywhere (code, commands, strings)
- **Data types** -- integers, floats, strings, booleans, arrays
- **String interpolation** -- `"hello @name"` just works, `'raw @string'` doesn't
- **Method syntax** -- `@name.upper()`, `@text.split(",")`, `@n.abs()`
- **`match-case`** -- pattern matching with `case _` wildcard
- **`const` / `pub`** -- immutability and export as language keywords
- **Shell variables** -- `@LAST_STATUS`, `@OS_NAME`, `@SHELL_PID` (read-only)
- **Procedures** -- first-class, closures, recursion
- **Loops** -- conditional (`loop (@i < 10)`) and iterator (`loop (@x : @items)`)
- **Command chaining** -- `&&`, `||`, pipes `|`, redirects `>`, `>>`
- **Background jobs** -- `command &`, `jobs`, `fg`, `wait`, `killjob`, `disown`
- **Output capture** -- `let @out = save(echo "hi")`
- **Scoped env** -- `with (@NODE_ENV = "prod") { ... }`
- **String/array/number methods** -- trim, replace, split, join, contains, slice, abs, etc.
- **Customizable prompt** -- oh-my-zsh style `{tokens}` with ANSI colors, git branch, time
- **Modern `ls`** -- icons, colors, grid/table layout, human-readable sizes, sorting
- **File operations** -- `mkdir`, `mkfile`, `rm`, `touch` as builtins
- **Directory stack** -- `pushdir`, `popdir`, `dirs`
- **Signal trapping** -- `trap 'echo bye' EXIT`
- **Interactive shell** -- history, tab completion (with cycling), `~/.dush/` config
- **Cross-platform** -- Windows, Linux, macOS

## Install

### Prerequisites
- Go 1.21+

### Build

```bash
# With gobake (recommended)
gobake build

# Or manually
go build -o build/dush ./cmd/dush
```

Binaries output to `build/`.

### Run

```bash
./build/dush
```

## Documentation

| Document | Description |
|---|---|
| [Language Reference](docs/language.md) | Variables, types, methods, control flow, procedures |
| [Built-in Commands](docs/commands.md) | All shell commands with flags and examples |
| [Shell Integration](docs/shell.md) | Pipes, redirects, background jobs, output capture |
| [Configuration](docs/configuration.md) | Prompt customization, config files, CLI flags |
| [Bundled Utilities](docs/utilities.md) | External tools: shout, ports, fmt, kill |
| [Language Spec](docs/LANGUAGE_SPEC.md) | Full formal language specification |

All built-in commands support `-h` for inline help.

## Project Structure

```
dush/
  cmd/
    dush/          # main shell binary
    shout/         # echo-like utility
    ports/         # cross-platform port scanner
    fmt/           # printf-style formatting
    kill/          # send signals to processes
  internal/
    parser/        # lexer, tokens, AST, Pratt parser
    evaluator/     # tree-walking evaluator, environment, methods, jobs
    prompt/        # customizable prompt renderer with ANSI colors
    builtins/      # built-in shell commands
    repl/          # interactive shell with history + tab completion
    app/           # global app state
    config/        # defaults and runtime config
    utils/         # utilities (history, colors, display)
  docs/            # documentation
  examples/
    showcase.dush  # comprehensive language demo
```

## License

MIT
