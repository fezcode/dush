# Shell Integration

## Commands

Any bare identifier that is not a keyword or known function is treated as a command:
```
git status
ls -la
echo "hello world"
atlas.ed file.txt          // dotted command names work
cd .config                 // dot-prefixed arguments work
```

## Command Arguments

- String literals: `echo "hello"`
- Raw strings: `echo 'hello @name'`
- Variables: `echo @name` (resolved)
- Grouped expressions: `echo (@x + 1)` (evaluated)
- Paths and flags: `echo file.txt`, `ls -la`, `cat --verbose`

## Command Chaining

```
echo "a" && echo "b"       // run next if previous succeeded
echo "a" || echo "b"       // run next if previous failed
ls | grep go               // pipe stdout to next command
```

## Redirects

```
echo "log" > output.txt    // redirect stdout (truncate)
echo "more" >> output.txt  // redirect stdout (append)
sort < input.txt           // redirect stdin from file
```

### Stderr Redirects

```
cmd 2> errors.txt          // redirect stderr (truncate)
cmd 2>> errors.txt         // redirect stderr (append)
cmd &> all.txt             // redirect both stdout+stderr to file
```

### Here-strings

Feed a string directly as stdin to a command:
```
sort <<< "banana\napple\ncherry"
cat <<< @message
```

## Background Jobs

```
sleep 30 &                 // [1] 12345
ping localhost &           // [2] 12346
ls | grep go &             // pipes work too
```

**Job management:**
```
jobs                       // list all background jobs
fg 1                       // wait for job 1 to finish
wait                       // wait for all jobs
killjob 1                  // kill a running job
disown 1                   // remove job from table
cleanjobs                  // remove finished entries
```

## Output Capture

```
let @output = save(echo "hello")
let @files = save(ls)
```

## Scoped Environment

```
with (@NODE_ENV = "production") {
    echo @NODE_ENV
}
```

## Source Files

Execute a script in the current environment:
```
source script.dush
. script.dush              // shorthand
```

## Interactive Keys

| Key        | Action                                       |
|------------|----------------------------------------------|
| `Tab`      | Complete / cycle through matches             |
| `Ctrl+R`   | Reverse incremental history search           |
| `Enter`    | (during Ctrl+R) accept the current match     |
| `Esc` / `Ctrl+G` / `Ctrl+C` | (during Ctrl+R) cancel         |
| `Backspace`| (during Ctrl+R) shorten the search pattern   |

Unterminated quotes (`'` or `"`) make dush prompt for a continuation line, the same way unclosed braces do.

## Signal Trapping

```
trap 'echo interrupted' INT
trap 'echo exiting' EXIT
trap '' INT                // remove trap
trap -l                   // list traps
```
