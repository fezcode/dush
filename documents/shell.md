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

## Scripts & Arguments

Invoke a `.dush` file directly:
```
dush build.dush prod --verbose
```

Arguments after the script filename land in `@ARGS` (array of strings), and
`@SCRIPT` holds the path of the running script.

```
// build.dush
echo "running: @SCRIPT"
echo "count:   @{@ARGS.len()}"
loop (@a : @ARGS) {
    echo "  - @a"
}
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

## Shell Modes

Three directives flip behavior-modifying flags for the current scope. They all share the same shape — `<name> on|off` with an optional block for scoped effect — and reset when the block or script exits.

### `strict on` — stop on first failure

Without strict, a failing command prints its error and the script keeps running (`@LAST_STATUS` records the non-zero exit). With strict, the first non-zero exit aborts the enclosing scope:

- `strict on` at top level → aborts the current script.
- `strict on { ... }` → aborts the block only; the outer script continues afterward. Use this to group "all or nothing" steps.

```
strict on {
    run_migration
    restart_service
    check_health
}
echo "strict group finished; LAST_STATUS=@LAST_STATUS"
```

Replaces bash's `set -e` without the historical footguns.

### `trace on` — echo each command before running it

Prints the command and its resolved args to stderr prefixed with `+ `, so you can see what the shell actually executed:

```
trace on {
    @name = "world"
    echo "hello @name"
}
// stderr will contain:
// + echo hello world
```

Replaces bash's `set -x`. Scoped form keeps the noise confined to the block you care about.

### `pipefail on` — propagate early-stage failures

Without pipefail, a pipeline's exit code is the exit code of its **last** stage:

```
producer | filter | sink
// @LAST_STATUS reflects sink only
```

That hides a failing `producer` if `sink` still succeeds. With pipefail, the pipeline's status becomes the **first** non-zero exit from any stage:

```
pipefail on
producer | filter | sink
// @LAST_STATUS is producer's exit code if producer failed
```

Particularly useful for `cmd | grep`, `curl | jq`, or any pipeline where the last stage tolerates empty input.

### Mixing the modes

```
strict on
pipefail on
trace on

run_deploy_step      // traced; aborts the script on any failure
```

Scoped and permanent forms stack naturally — a block form always takes effect over whatever the enclosing scope had and restores it on exit.

## Signal Trapping

```
trap 'echo interrupted' INT
trap 'echo exiting' EXIT
trap '' INT                // remove trap
trap -l                   // list traps
```
