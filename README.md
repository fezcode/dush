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
- **Customizable prompt** -- oh-my-zsh style `{tokens}` with ANSI colors, git branch, time
- **Interactive shell** -- history, tab completion, `~/.dush/` config
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

## CLI Usage

```
dush                          # interactive mode
dush script.dush              # run a script (non-interactive)
dush -v / --version           # print version
dush -h / --help              # print help
```

### Path override flags

```
dush --env <path>             # override env file (default: ~/.dush/env)
dush --is <path>              # override is file (default: ~/.dush/is)
dush --history <path>         # override history file (default: ~/.dush/history)
```

Useful for testing or running isolated sessions:
```bash
dush --env ./test/env --is /dev/null --history /tmp/h examples/showcase.dush
```

## Configuration

All config lives in `~/.dush/`:

| File | Loaded | Purpose |
|---|---|---|
| `~/.dush/env` | Always (interactive + scripts) | Environment setup, pub vars, prompt config |
| `~/.dush/is` | Interactive only | Aliases, greeting, shell helpers |
| `~/.dush/history` | Interactive only | Command history (auto-managed) |

**Loading order:** `env` first, then `is`.

### Non-interactive mode

`dush script.dush` only sources `~/.dush/env`, skipping `~/.dush/is`. Scripts run fast without interactive setup.

## Prompt

The prompt is fully customizable via `@PROMPT_LINE` in `~/.dush/env`. It uses `{token}` syntax inspired by oh-my-zsh.

### Default prompt

```
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} {fg:yellow}${reset} '
```

### Content tokens

| Token | Output |
|---|---|
| `{user}` | Current username |
| `{host}` | Machine hostname |
| `{dir}` | Current directory basename |
| `{path}` | Full absolute path |
| `{home_path}` | Path with `~` replacing home dir |
| `{time}` | Current time `15:04:05` |
| `{date}` | Current date `2006-01-02` |
| `{git}` | Git branch name (empty if not in a repo) |
| `{last_status}` | Exit code of the last command |
| `{os}` | Operating system (`windows`, `linux`, `darwin`) |
| `{newline}` | Line break (for multi-line prompts) |

### Color and style tokens

| Token | Effect |
|---|---|
| `{fg:red}` | Set foreground to a named color |
| `{bg:blue}` | Set background to a named color |
| `{fg:#ff5500}` | Set foreground to a hex color (true-color) |
| `{bg:#1a1a2e}` | Set background to a hex color (true-color) |
| `{reset}` | Reset all colors and styles |
| `{bold}` | Bold text |
| `{dim}` | Dim/faint text |
| `{italic}` | Italic text |
| `{underline}` | Underlined text |

**Named colors:** `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` -- plus `bright_black` through `bright_white`.

### Continuation prompt

For multi-line input (open braces, etc.):

```
@CONTINUATION_PROMPT = '{fg:yellow}...{reset} '
```

### Example prompts

```bash
// Minimal
@PROMPT_LINE = '$ '

// Classic user@host:path
@PROMPT_LINE = '{fg:green}{user}{reset}@{host}:{fg:blue}{home_path}{reset}$ '

// With git branch
@PROMPT_LINE = '{fg:cyan}{user}{reset}:{fg:blue}{home_path}{reset} ({fg:magenta}{git}{reset}) $ '

// Two-line with timestamp
@PROMPT_LINE = '{fg:dim}{time}{reset} {fg:cyan}{user}{reset}@{host}:{fg:yellow}{home_path}{reset}{newline}{fg:green}${reset} '

// Powerline-style with hex colors
@PROMPT_LINE = '{bg:#1a1a2e}{fg:#ff6b6b} {user} {reset}{bg:#16213e}{fg:#a8d8ea} {home_path} {reset}{fg:yellow} > {reset}'

// Show last exit code when non-zero
@PROMPT_LINE = '{fg:red}{last_status}{reset} {fg:cyan}{dir}{reset} $ '
```

### Example `~/.dush/env`

```bash
pub @EDITOR = "vim"
pub @LANG = "en_US.UTF-8"
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} $ '
```

### Example `~/.dush/is`

```bash
alias ll='ls -la'
alias gs='git status'
echo "Welcome back, @USER_NAME!"
```

## License

MIT
