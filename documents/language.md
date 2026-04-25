# Dush Language Reference

Dush is a strictly typed, C-style scripting language that treats shell commands as first-class citizens. Variables use the `@` sigil for unambiguous access everywhere -- in code, commands, and strings.

## Syntax

- **Blocks:** `{}`
- **Statement terminators:** Newlines. Semicolons are optional.
- **Comments:** `//` single-line comments.

## Variables

### Declaration & Assignment

```
@x = 10                    // create or update
let @x = 10               // explicit local scope (shadows parent)
const @PI = 3.14           // immutable
pub @KEY = "value"         // exported to child processes
pub const @VER = "1.0"     // exported + immutable
```

### Access

Variables always use the `@` sigil: `@x`, `@name`, `@LAST_STATUS`.

Bare identifiers in commands are always string literals:
```
echo hello                 // prints "hello" (literal)
echo @name                 // prints the value of @name
```

## Data Types

### Integer

```
@x = 42
@y = -10
@z = @x + @y              // 32
@z = 15 % 4               // 3

@n = -5
@n.abs()                   // 5
@n.to_string()             // "-5"
```

### Float

Mixed int/float operations produce floats.

```
@pi = 3.14
@x = 1 + 2.5              // 3.5
@y = 7.0 / 2.0            // 3.5

@f = -3.7
@f.abs()                   // 3.7
@f.to_string()             // "-3.7"
```

### String

Double-quoted strings support `@` interpolation and backslash escapes. Single-quoted strings are raw (no interpolation, no escapes).

```
@name = "world"
@greeting = "hello @name"  // "hello world"
@raw = 'hello @name'       // "hello @name"
@concat = "hello" + " " + "world"
@ch = @name[0]             // "w" (index access)
@last = @name[-1]          // "d" (negative index)
```

**Escape sequences** (double-quoted only):

| Escape | Result |
|---|---|
| `\n` | Newline |
| `\t` | Tab |
| `\r` | Carriage return |
| `\0` | NUL byte |
| `\\` | Backslash |
| `\"` | Double quote |
| `\'` | Single quote |
| `\@` | Literal `@` (suppresses interpolation) |
| `\xHH` | Hex byte (2 digits) |
| `\u{HHHH}` | Unicode codepoint (1–6 hex digits) |

```
echo "line 1\nline 2"           // two lines
echo "\x41\x42"                  // "AB"
echo "\u{1F600}"                 // 😀
echo "email: \@example.com"      // literal @
```

**Inline expressions** with `@{...}` — inner `"` does not need escaping:

```
@items = ["a", "b", "c"]
echo "joined: @{@items.join(", ")}"          // joined: a, b, c
echo "branch: @{save(git rev-parse --abbrev-ref HEAD).chomp()}"
```

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `.upper()` | String | Uppercase |
| `.lower()` | String | Lowercase |
| `.len()` | Integer | Character count |
| `.trim()` | String | Strip whitespace |
| `.trim_start()` | String | Strip leading whitespace |
| `.trim_end()` | String | Strip trailing whitespace |
| `.contains(s)` | Boolean | Substring check |
| `.starts_with(s)` | Boolean | Prefix check |
| `.ends_with(s)` | Boolean | Suffix check |
| `.replace(old, new)` | String | Replace first occurrence |
| `.replace_all(old, new)` | String | Replace all occurrences (literal) |
| `.replace_regex(pat, new)` | String | Replace all matches of regex `pat` |
| `.match(pat)` | Array | Regex match: `[whole, cap1, cap2, ...]` or `[]` if no match |
| `.matches(pat)` | Boolean | Regex match test (true/false) |
| `.match_all(pat)` | Array | All non-overlapping matches as strings |
| `.split(sep)` | Array | Split into array |
| `.slice(start, end)` | String | Substring by index |
| `.chomp()` | String | Strip one trailing `\n` or `\r\n` (handy for `save()` output) |
| `.or(default)` | String | Return default if empty |
| `.to_string()` | String | Identity |

### Boolean

```
@flag = true
@other = false
@x = 1 < 2                // true
@y = !@flag                // false
```

### Array

```
@arr = [1, 2, 3, 4, 5]
@names = ["alice", "bob", "charlie"]
@mixed = [1, "two", true, 3.14]
@empty = []
```

**Index access** (0-based, negative wraps from end):
```
@arr[0]                    // 1
@arr[-1]                   // 5
@arr[2] = 99               // assignment
```

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `.len()` | Integer | Element count |
| `.join(sep)` | String | Join elements with separator |
| `.contains(val)` | Boolean | Check if value exists |
| `.push(val...)` | Array | Append elements (mutates) |
| `.pop()` | any | Remove and return last element |
| `.first()` | any | First element (or null) |
| `.last()` | any | Last element (or null) |
| `.slice(start, end?)` | Array | Sub-array by index |
| `.reverse()` | Array | Reversed copy |
| `.map(fn)` | Array | Transform each element |
| `.filter(fn)` | Array | Keep elements where fn returns true |

```
@nums = [1, 2, 3, 4, 5]
@doubled = @nums.map(proc(@x) { return @x * 2 })       // [2, 4, 6, 8, 10]
@evens = @nums.filter(proc(@x) { return @x % 2 == 0 })  // [2, 4]
@nums.push(6)                                             // [1, 2, 3, 4, 5, 6]
echo @nums.join(", ")                                     // 1, 2, 3, 4, 5, 6
```

Looping:
```
loop (@item : @arr) {
    echo @item
}
```

### Map

```
@user = {"name": "alice", "age": 30, "active": true}
@empty = {}
```

**Key access and assignment:**
```
@user["name"]              // "alice"
@user["age"] = 31          // update
@user["email"] = "a@b.c"  // add new key
```

**Methods:**

| Method | Returns | Description |
|---|---|---|
| `.len()` | Integer | Number of pairs |
| `.keys()` | Array | All keys in insertion order |
| `.values()` | Array | All values in insertion order |
| `.has(key)` | Boolean | Check if key exists |
| `.delete(key)` | Boolean | Remove a key, returns true if existed |
| `.merge(other)` | Map | New map with both maps' pairs (other wins conflicts) |

```
@config = {"host": "localhost", "port": "8080"}
echo @config.keys().join(", ")          // host, port
echo @config.has("host")                // true
@config.delete("port")

@extra = {"debug": "true"}
@merged = @config.merge(@extra)         // host + debug
```

Looping over map keys:
```
loop (@key : @config) {
    echo @key
}
```

### Type Conversion

```
int(3.7)                   // 3
int("42")                  // 42
int(true)                  // 1
float(3)                   // 3.0
float("3.14")              // 3.14
type(42)                   // "INTEGER"
type("hello")              // "STRING"
```

## Methods

Methods can be chained:

```
@s = "  HELLO WORLD  "
@s.trim().lower()                    // "hello world"
@s.trim().split(" ").join("-")       // "HELLO-WORLD"
```

## Control Flow

### If / Else

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

Boolean condition pattern:
```
match (true) {
    case (@temp < 32) { echo "freezing" }
    case (@temp < 80) { echo "pleasant" }
    case _ { echo "hot" }
}
```

### Loops

```
loop (@x < 10) { ... }              // conditional
loop (@item : @arr) { ... }         // iterator (array)
loop (@i : 5) { ... }               // iterator (range 0..4)
```

## Procedures

```
proc greet(@name) {
    echo "hello @name"
}
greet("world")
```

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
@count()    // 1
@count()    // 2
```

## Operators

### Arithmetic
`+` (add/concat), `-`, `*`, `/`, `%`

### Comparison
`==`, `!=`, `<`, `>`

### Logical
`&&` (short-circuit AND), `||` (short-circuit OR), `!` (NOT)

### Shell
`|` (pipe), `>` (redirect truncate), `>>` (redirect append), `<` (stdin from file), `&` (background)

## Shell Modes

Three directives toggle behavior flags for the current scope. Each takes `on` or `off`, with an optional block for scoped effect. See [shell.md → Shell Modes](shell.md) for full details.

```
strict on               // abort current scope on first non-zero exit
trace on                // print each command before running it
pipefail on             // pipeline status = first non-zero stage (not last)

strict off              // inverse
```

Block form applies the mode only inside, and restores the prior value on exit:

```
strict on {
    run_migration
    restart_service      // block aborts here if run_migration fails,
}                         // but the outer script keeps going
```

## Shell Variables (read-only)

| Variable | Description |
|---|---|
| `@LAST_STATUS` | Exit code of last command |
| `@SHELL_PID` | Process ID of the shell |
| `@WORKING_DIR` | Current working directory |
| `@HOME_DIR` | User's home directory |
| `@OS_NAME` | Operating system name |
| `@USER_NAME` | Current user |
| `@SHELL_VERSION` | Dush version string |
| `@SCRIPT` | Path of the currently-running script (empty in interactive mode) |
| `@ARGS` | Array of arguments passed to the script after its filename |

## Built-in Functions

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

## Roadmap

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
- JSON parse/stringify builtins
- HTTP fetch builtin
