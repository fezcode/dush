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

## 4. Data Types

### Integer
Whole numbers. Supports standard arithmetic: `+`, `-`, `*`, `/`, `%`.

```
@x = 42
@y = -10
@z = @x + @y        # 32
@z = 15 % 4         # 3

// Methods
@n = -5
@n.abs()             # 5
@n.to_string()       # "-5"
```

### Float
Decimal numbers. Mixed int/float operations produce floats.

```
@pi = 3.14
@x = 1 + 2.5        # 3.5 (int + float = float)
@y = 7.0 / 2.0      # 3.5

// Methods
@f = -3.7
@f.abs()             # 3.7
@f.to_string()       # "-3.7"
```

### String
Double-quoted strings support `@` interpolation and backslash escapes (`\n`, `\t`, `\r`, `\0`, `\\`, `\"`, `\'`, `\@`, `\xHH`, `\u{HHHH}`). Single-quoted strings are raw (no interpolation, no escapes).

Inside `@{...}` interpolation blocks, bare `"` is literal — no escaping needed.

```
@name = "world"
@greeting = "hello @name"       # "hello world"
@raw = 'hello @name'            # "hello @name" (literal)
@multi = "line1\nline2"         # two lines
@inline = "sum: @{1 + 2}"       # "sum: 3"
@items = ["a", "b"]
@csv = "list: @{@items.join(", ")}"
@concat = "hello" + " " + "world"
```

Comparison: `==`, `!=`, `<`, `>` (lexicographic).

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `.upper()` | String | Uppercase |
| `.lower()` | String | Lowercase |
| `.len()` | Integer | Character count |
| `.trim()` | String | Strip leading/trailing whitespace |
| `.trim_start()` | String | Strip leading whitespace |
| `.trim_end()` | String | Strip trailing whitespace |
| `.chomp()` | String | Strip one trailing `\n` or `\r\n` |
| `.contains(s)` | Boolean | Substring check |
| `.starts_with(s)` | Boolean | Prefix check |
| `.ends_with(s)` | Boolean | Suffix check |
| `.replace(old, new)` | String | Replace first occurrence |
| `.replace_all(old, new)` | String | Replace all occurrences |
| `.split(sep)` | Array | Split into array |
| `.slice(start, end)` | String | Substring by index |
| `.or(default)` | String | Return default if empty |
| `.to_string()` | String | Identity |

### Boolean
```
@flag = true
@other = false
@x = 1 < 2          # true
@y = !@flag          # false
```

### Array
Arrays are created by functions like `split()`. They hold ordered elements.

```
@arr = split("a,b,c", ",")
@arr.len()                   # 3
@arr.join(" | ")             # "a | b | c"
@arr.contains("b")          # true

loop (@item : @arr) {
    echo @item
}
```

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `.len()` | Integer | Element count |
| `.join(sep)` | String | Join elements with separator |
| `.contains(val)` | Boolean | Membership check |

### Type Conversion
```
int(3.7)             # 3
int("42")            # 42
int(true)            # 1
int(false)           # 0

float(3)             # 3.0
float("3.14")        # 3.14

type(42)             # "INTEGER"
type("hello")        # "STRING"
type(true)           # "BOOLEAN"
type(3.14)           # "FLOAT"
```

## 5. Methods

Variables support method syntax. Methods can be chained:

```
@s = "  HELLO WORLD  "
@s.trim().lower()            # "hello world"
@s.trim().split(" ").join("-")  # "HELLO-WORLD"
```

See the per-type method tables in Section 4 for the full list.

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
- **Recursion:** procedures can call themselves
- **Closures:** inner procs capture the enclosing environment

```
proc make_counter() {
    @n = 0
    return proc() {
        @n = @n + 1
        return @n
    }
}
let @count = make_counter()
@count()    # 1
@count()    # 2
```

## 8. Operators

### Arithmetic
| Operator | Description | Example |
|---|---|---|
| `+` | Add / string concat | `2 + 3` → `5` |
| `-` | Subtract | `10 - 4` → `6` |
| `*` | Multiply | `3 * 4` → `12` |
| `/` | Divide | `10 / 3` → `3` |
| `%` | Modulo | `10 % 3` → `1` |

### Comparison
| Operator | Description |
|---|---|
| `==` | Equal |
| `!=` | Not equal |
| `<` | Less than |
| `>` | Greater than |

### Logical
| Operator | Description |
|---|---|
| `&&` | Logical AND (short-circuit) |
| `\|\|` | Logical OR (short-circuit) |
| `!` | Logical NOT (prefix) |

### Shell
| Operator | Description |
|---|---|
| `\|` | Pipe stdout to next command's stdin |
| `>` | Redirect stdout (truncate) |
| `>>` | Redirect stdout (append) |
| `<` | Redirect stdin from file |
| `&` | Run command in background |

## 9. Shell Integration

### Commands
Any bare identifier that is not a keyword or known function is treated as a command:
```
git status
ls -la
echo "hello world"
atlas.ed file.txt              # dotted command names work
cd .config                     # dot-prefixed arguments work
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

### Background jobs
```
sleep 30 &                     # [1] 12345
jobs                           # list all background jobs
fg 1                           # wait for job 1 to finish
killjob 1                      # terminate job 1
cleanjobs                      # remove finished jobs from list
```

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

## 10. Shell Variables (read-only)

These are populated by the shell and cannot be reassigned:
- `@LAST_STATUS` — exit code of last command
- `@SHELL_PID` — process ID of the shell
- `@WORKING_DIR` — current working directory
- `@HOME_DIR` — user's home directory
- `@OS_NAME` — operating system name
- `@USER_NAME` — current user
- `@SHELL_VERSION` — dush version string
- `@SCRIPT` — path of the running script (empty in interactive mode)
- `@ARGS` — array of arguments passed to the script

## 10a. Shell Modes

Three directives toggle language-level behavior for the current scope. Each takes `on` or `off`, with an optional block form that scopes the effect and reverts on exit.

| Directive     | Effect                                                                  |
|---------------|-------------------------------------------------------------------------|
| `strict on`   | Abort the enclosing scope (script or block) on the first non-zero exit |
| `trace on`    | Print each command with resolved args to stderr, prefixed `+ `         |
| `pipefail on` | Pipeline status = first non-zero stage instead of the last stage's     |

```
strict on              # rest of script: stop on first failure
trace on { cmd }       # block only
pipefail off           # inverse
```

Scoped-block strict aborts the block but does **not** propagate out — the outer script keeps going. Use `LAST_STATUS` to check after the block.

## 11. Built-in Functions

| Function | Description |
|---|---|
| `len(val)` | String or array length |
| `split(str, sep)` | Split string into array |
| `join(arr, sep)` | Join array into string |
| `replace(str, old, new)` | Replace all occurrences |
| `to_upper(str)` | Uppercase string |
| `to_lower(str)` | Lowercase string |
| `trim(str)` | Strip whitespace |
| `contains(str, sub)` | Substring check |
| `format(fmt, ...)` | Printf-style formatting |
| `type(val)` | Type name as string |
| `int(val)` | Convert to integer |
| `float(val)` | Convert to float |
| `save(cmd)` | Capture command output |
| `exists(path)` | File/directory exists check |
| `is_dir(path)` | Directory check |

## 12. Built-in Shell Commands

| Command | Description |
|---|---|
| `cd [path]` | Change directory |
| `pwd` | Print working directory |
| `ls [opts] [path]` | List directory (modern, with icons) |
| `echo [args...]` | Print arguments |
| `clear` | Clear screen |
| `alias name='value'` | Define alias |
| `unalias name` | Remove alias |
| `export @VAR = value` | Export variable |
| `history` | Show command history |
| `sleep <seconds>` | Sleep |
| `help` | Show built-in help |
| `version` | Show version |
| `jobs` | List background jobs |
| `fg <id>` | Wait for background job |
| `killjob <id>` | Kill background job |
| `cleanjobs` | Remove finished jobs |

### `ls` options

| Flag | Description |
|---|---|
| `-l` / `--long` | Long table format with permissions, size, owner, timestamps |
| `-a` / `--all` | Show hidden (dot) files |
| `-1` / `--oneline` | One entry per line |
| `-r` / `--reverse` | Reverse sort order |
| `-S` / `--sort=size` | Sort by file size |
| `-t` / `--sort=time` | Sort by modification time |
| `-X` / `--sort=ext` | Sort by file extension |
| `--no-header` | Hide table header in long format |
| `--no-icons` | Hide file type icons |

## 13. Configuration & Startup

All config lives in `~/.dush/`:

| File | Loaded | Purpose |
|---|---|---|
| `env` | Always (interactive + scripts) | Environment setup, pub vars, prompt config |
| `is` | Interactive only | Aliases, greeting, interactive helpers |
| `history` | Interactive only | Command history (auto-managed) |

### Loading Order
1. `~/.dush/env` — **always loaded** (interactive + scripts). Environment setup, pub vars, prompt config.
2. `~/.dush/is` — **interactive only**. Aliases, greeting, interactive helpers.

### CLI Flags
```
dush --env <path>       # override env file path
dush --is <path>        # override is file path
dush --history <path>   # override history file path
```

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

## 14. Roadmap

Proposed syntax is dush-native — no `$1` / `$(...)` / `set -e` / `!!` borrowings.

**Error handling:**
```
try {
    risky()
} catch (@err) {
    echo "failed: @err"
}
throw "bad state"
```

Other items still on the list:
- Here-docs (`<<EOF ... EOF`)
- Regex match with captures
- JSON parse/stringify builtins
- HTTP fetch builtin
