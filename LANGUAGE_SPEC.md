# Dush Language Specification (v2)

This document defines the syntax and semantics of the Dush scripting language.

## 1. Core Philosophy
Dush is a strictly typed, C-style scripting language that treats shell commands as first-class citizens. Variables use the `@` sigil for unambiguous access everywhere — in code, commands, and strings.

## 2. Basic Syntax
- **Blocks:** Defined by curly braces `{}`.
- **Statement Terminators:** New-lines. Semicolons are optional.
- **Comments:** Single-line comments start with `//`.

## 3. Variables

### Declaration & Assignment
- **Simple assignment:** `@x = 10` (creates or updates variable)
- **Let (explicit local scope):** `let @x = 10` (shadows parent scope)
- **Constant:** `const @PI = 3.14` (immutable, cannot be reassigned)
- **Public/Export:** `pub @KEY = "value"` (exported to child processes)
- **Public constant:** `pub const @VER = "1.0"`

### Access
Variables always use the `@` sigil: `@x`, `@name`, `@LAST_STATUS`.

Bare identifiers (without `@`) in commands are always treated as string literals:
- `echo hello` → prints "hello" (literal string)
- `echo @name` → prints the value of variable `name`

### Data Types
- `int`: `@x = 42`
- `float`: `@pi = 3.14`
- `string`: `@name = "world"`
- `bool`: `@flag = true`
- `array`: from `split()` and other functions

## 4. Strings

### Double-quoted strings (interpolation)
Variables inside double-quoted strings are automatically interpolated:
```
@name = "world"
echo "hello @name"    → hello world
```

### Single-quoted strings (raw)
Single-quoted strings have no interpolation:
```
echo 'hello @name'    → hello @name
```

## 5. Methods

Variables support method syntax for common operations:

### String methods
- `@s.upper()` — uppercase
- `@s.lower()` — lowercase
- `@s.len()` — length (returns int)
- `@s.trim()`, `@s.trim_start()`, `@s.trim_end()` — whitespace trimming
- `@s.contains("sub")` — substring check (returns bool)
- `@s.starts_with("pre")`, `@s.ends_with("suf")` — prefix/suffix check
- `@s.replace("old", "new")` — replace first occurrence
- `@s.replace_all("old", "new")` — replace all occurrences
- `@s.split(" ")` — split into array
- `@s.slice(start, end)` — substring
- `@s.or("default")` — return default if empty
- `@s.to_string()` — identity (already a string)

### Array methods
- `@arr.len()` — length
- `@arr.join(",")` — join elements
- `@arr.contains("item")` — membership check

### Number methods
- `@n.abs()` — absolute value
- `@n.to_string()` — convert to string

## 6. Control Flow

### Conditionals
```
if (@x > 0) {
    echo "positive"
} else {
    echo "non-positive"
}
```

### Match / Case
```
match (@value) {
    case 1 { echo "one" }
    case "hello" { echo "greeting" }
    case _ { echo "default" }
}
```

`case _` is the wildcard/default branch. Cases are evaluated top-to-bottom; the first match wins. Supports integers, floats, strings, and booleans as case values.

Boolean condition pattern:
```
match (true) {
    case (@temp < 32) { echo "freezing" }
    case (@temp < 80) { echo "pleasant" }
    case _ { echo "hot" }
}
```

### Loops
- **Conditional:** `loop (@x < 10) { ... }`
- **Iterator (array):** `loop (@item : @arr) { ... }`
- **Iterator (range):** `loop (@i : 5) { ... }` (iterates 0..4)

## 7. Procedures
```
proc greet(@name) {
    echo "hello @name"
}
greet("world")
```

- **Declaration:** `proc name(@param1, @param2) { ... }`
- **Proc literal:** `let @add = proc(@x, @y) { @x + @y }`
- **Return:** `return @value`

## 8. Shell Integration

### Commands
Any bare identifier that is not a keyword or known function is treated as a command:
```
git status
ls -la
echo "hello world"
```

### Command arguments
- String literals: `echo "hello"`
- Raw strings: `echo 'hello @name'`
- Variables: `echo @name` (resolved)
- Grouped expressions: `echo (@x + 1)` (evaluated)
- Paths and flags: `echo file.txt`, `ls -la`, `cat --verbose`

### Command chaining
- `&&`: run next if previous succeeded (exit code 0)
- `||`: run next if previous failed
- `|`: pipe stdout to next command's stdin

### Output capture
```
let @output = save(echo "hello")
```

### Inline environment
```
with (@NODE_ENV = "production") {
    echo @NODE_ENV
}
```

## 9. Shell Variables (read-only)

These are populated by the shell and cannot be reassigned:
- `@LAST_STATUS` — exit code of last command
- `@SHELL_PID` — process ID of the shell
- `@WORKING_DIR` — current working directory
- `@HOME_DIR` — user's home directory
- `@OS_NAME` — operating system name
- `@USER_NAME` — current user
- `@SHELL_VERSION` — dush version string

## 10. Built-in Functions
- `len(str)` — string/array length
- `split(str, sep)` — split string into array
- `replace(str, old, new)` — replace in string
- `to_upper(str)` — uppercase string
- `to_lower(str)` — lowercase string
- `type(val)` — type name as string
- `save(cmd)` — capture command output
- `exists(path)` — file/directory exists check
- `is_dir(path)` — directory check

## 11. Configuration & Startup

All config lives in `~/.dush/`:

| File | Loaded | Purpose |
|---|---|---|
| `env` | Always (interactive + scripts) | Environment setup, pub vars, prompt config |
| `is` | Interactive only | Aliases, greeting, interactive helpers |
| `history` | Interactive only | Command history (auto-managed) |

### Loading Order
1. `~/.dush/env` — **always loaded** (interactive + scripts). Environment setup, pub vars, prompt config.
2. `~/.dush/is` — **interactive only**. Aliases, greeting, interactive helpers.

### Prompt Configuration
Set `@PROMPT_LINE` in `~/.dush/env` to customize the prompt using tokens:

```
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} {fg:yellow}${reset} '
```

#### Available tokens

| Token | Description |
|---|---|
| `{user}` | Current username |
| `{host}` | Hostname |
| `{dir}` | Current directory name (basename) |
| `{path}` | Full absolute path |
| `{home_path}` | Path with `~` for home directory |
| `{time}` | Current time (HH:MM:SS) |
| `{date}` | Current date (YYYY-MM-DD) |
| `{git}` | Current git branch (empty if not in repo) |
| `{last_status}` | Exit code of last command |
| `{os}` | Operating system name |
| `{newline}` | Line break |

#### Color tokens

| Token | Description |
|---|---|
| `{fg:COLOR}` | Foreground color |
| `{bg:COLOR}` | Background color |
| `{reset}` | Reset all colors/styles |
| `{bold}` | Bold text |
| `{dim}` | Dim text |
| `{italic}` | Italic text |
| `{underline}` | Underlined text |

Named colors: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, plus `bright_*` variants.
Hex colors: `{fg:#ff5500}`, `{bg:#1a1a2e}`.

#### Continuation prompt
Set `@CONTINUATION_PROMPT` to customize the multi-line input prompt (default: `{fg:yellow}...{reset} `).

#### Example prompts
```
// Minimal
@PROMPT_LINE = '$ '

// With git branch
@PROMPT_LINE = '{fg:green}{user}{reset}:{fg:blue}{home_path}{reset} ({fg:magenta}{git}{reset}) $ '

// Two-line with time
@PROMPT_LINE = '{fg:dim}{time}{reset} {fg:cyan}{user}{reset}@{fg:cyan}{host}{reset}:{fg:yellow}{home_path}{reset}{newline}{fg:green}${reset} '

// Colorful with status indicator
@PROMPT_LINE = '{bg:#1a1a2e}{fg:#ff6b6b} {user} {reset}{fg:blue} {home_path} {reset}{fg:yellow}>{reset} '
```

### Non-interactive Mode
Running `dush script.dush` only sources `~/.dush/env`, skipping `~/.dush/is`. This keeps script execution fast and predictable.

### Example `~/.dush/env`
```
pub @EDITOR = "vim"
pub @LANG = "en_US.UTF-8"
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} $ '
```

### Example `~/.dush/is`
```
alias ll='ls -la'
alias gs='git status'
echo "Welcome back, @USER_NAME!"
```

## 12. Roadmap
- Pipes & redirection: `ls | grep "go"`, `echo "hello" > file.txt`
- Globbing: `ls *.go`
- Map type: `@m = {"key": "value"}`
- Error handling patterns
