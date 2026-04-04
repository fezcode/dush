# Dush

![Dush Logo](logo.svg)

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
```

## Features

- **`@` variable system** -- unambiguous access everywhere (code, commands, strings)
- **String interpolation** -- `"hello @name"` just works, `'raw @string'` doesn't
- **Method syntax** -- `@name.upper()`, `@text.split(",")`, `@n.abs()`
- **`match-case`** -- pattern matching with `case _` wildcard
- **`const` / `pub`** -- immutability and export as language keywords
- **Shell variables** -- `@LAST_STATUS`, `@OS_NAME`, `@SHELL_PID` (read-only)
- **Procedures** -- first-class, closures, recursion
- **Loops** -- conditional (`loop (@i < 10)`) and iterator (`loop (@x : @items)`)
- **Command chaining** -- `&&`, `||`, pipes `|`, redirects `>`, `>>`
- **Output capture** -- `let @out = save(echo "hi")`
- **Scoped env** -- `with (@NODE_ENV = "prod") { ... }`
- **String/array/number methods** -- trim, replace, split, join, contains, slice, abs, etc.
- **Interactive shell** -- history, tab completion, `~/.dushis` profile
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

## The `@` System

Every variable uses `@`. No exceptions, no ambiguity.

```bash
@x = 10                        # assign
let @x = 10                    # explicit local scope
const @MAX = 100               # immutable
pub @KEY = "secret"            # exported to child processes
pub const @VER = "2.0"         # exported + immutable
```

Bare identifiers are always string literals in commands:
```bash
echo hello                     # prints "hello" (literal)
echo @name                     # prints value of name
```

### String Interpolation

```bash
@lang = "dush"
echo "welcome to @lang"        # welcome to dush
echo 'no @interpolation'       # no @interpolation
```

### Methods

```bash
@s = "hello world"
@s.upper()                     # HELLO WORLD
@s.split(" ")                  # [hello, world]
@s.len()                       # 11
@s.contains("world")           # true
@s.replace("world", "dush")   # hello dush
@s.slice(0, 5)                 # hello

@arr = split("a,b,c", ",")
@arr.join(" | ")               # a | b | c
@arr.len()                     # 3

@n = -42
@n.abs()                       # 42
```

### Shell Variables (read-only)

```bash
@LAST_STATUS    # exit code of last command
@OS_NAME        # windows / linux / darwin
@USER_NAME      # current user
@HOME_DIR       # home directory
@WORKING_DIR    # current working directory
@SHELL_PID      # process ID
@SHELL_VERSION  # dush version
```

## Control Flow

### If / Else

```bash
if (@score > 90) {
    echo "A"
} else {
    echo "not A"
}
```

### Match / Case

```bash
match (@code) {
    case 0 { echo "success" }
    case 1 { echo "error" }
    case _ { echo "unknown" }
}
```

### Loops

```bash
# Conditional
@i = 0
loop (@i < 10) {
    @i = @i + 1
}

# Iterator (range)
loop (@i : 5) { echo @i }         # 0 1 2 3 4

# Iterator (array)
loop (@item : @items) { echo @item }
```

## Procedures

```bash
proc add(@a, @b) {
    return @a + @b
}
add(2, 3)    # 5

# First-class
let @double = proc(@x) { @x * 2 }
@double(21)  # 42

# Closures
proc make_adder(@n) {
    return proc(@x) { @x + @n }
}
let @add5 = make_adder(5)
@add5(10)    # 15
```

## Shell Integration

```bash
# Commands
git status
ls -la

# Chaining
echo "a" && echo "b"          # both run
echo "a" || echo "b"          # only first runs

# Output capture
let @files = save(ls)

# Scoped environment
with (@NODE_ENV = "production") {
    echo @NODE_ENV
}
```

## Project Structure

```
dush/
  cmd/
    dush/          # main shell binary
    shout/         # echo-like utility
    ports/         # cross-platform port scanner
  internal/
    parser/        # lexer, tokens, AST, Pratt parser
    evaluator/     # tree-walking evaluator, environment, methods
    prompt/        # customizable prompt renderer with ANSI colors
    builtins/      # built-in shell commands (cd, pwd, help)
    repl/          # interactive shell with history + tab completion
    app/           # global app state
    config/        # defaults and runtime config
    utils/         # utilities (history, display)
  examples/
    showcase.dush  # comprehensive language demo
  Recipe.go        # gobake build recipe
  LANGUAGE_SPEC.md # full language specification
```

## Bundled Utilities

- **shout** -- like `echo`, but ours
- **ports** -- cross-platform open port scanner (Windows/Linux/macOS native APIs)

## Configuration

All config lives in `~/.dush/`:

| File | Loaded | Purpose |
|---|---|---|
| `~/.dush/env` | Always (interactive + scripts) | Environment setup, pub vars, prompt config |
| `~/.dush/is` | Interactive only | Aliases, greeting, shell helpers |
| `~/.dush/history` | Interactive only | Command history (auto-managed) |

**Loading order:** `env` first, then `is`.

### Prompt customization (in `~/.dush/env`)

```bash
// Default
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} {fg:yellow}${reset} '

// With git branch and time
@PROMPT_LINE = '{fg:dim}{time}{reset} {fg:green}{user}{reset}:{fg:blue}{home_path}{reset} ({fg:magenta}{git}{reset}) $ '

// Hex colors
@PROMPT_LINE = '{fg:#ff6b6b}{user}{reset} {fg:#a8d8ea}{home_path}{reset} > '
```

**Tokens:** `{user}`, `{host}`, `{dir}`, `{path}`, `{home_path}`, `{time}`, `{date}`, `{git}`, `{last_status}`, `{os}`, `{newline}`
**Colors:** `{fg:red}`, `{bg:blue}`, `{fg:#ff5500}`, `{reset}`, `{bold}`, `{dim}`, `{italic}`, `{underline}`
**Named colors:** black, red, green, yellow, blue, magenta, cyan, white + bright_* variants

### Non-interactive mode

`dush script.dush` only sources `~/.dush/env`, skipping `~/.dush/is`. Scripts run fast without interactive setup.

## License

MIT
