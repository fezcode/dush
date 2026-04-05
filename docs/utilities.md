# Bundled Utilities

Dush ships with standalone command-line utilities in `cmd/`. These are built as separate binaries alongside the main shell.

## shout

An `echo`-like utility.

```
shout hello world           # hello world
shout -n "no newline"       # no newline (no trailing \n)
```

## ports

Cross-platform open port scanner using native OS APIs (Windows/Linux/macOS).

```
ports                       # list all open ports
```

## fmt

Printf-style formatting utility.

```
fmt "%s is %d years old\n" Alice 30
fmt "hex: %x\n" 255
fmt "pi is %.2f\n" 3.14159
```

**Format specifiers:** `%s` (string), `%d` (integer), `%f` (float), `%x` (hex), `%o` (octal), `%b` (binary), `%q` (quoted), `%v` (default), `%%` (literal percent).

**Escape sequences:** `\n` (newline), `\t` (tab), `\r` (carriage return).

Arguments are auto-detected as integer, float, or string.

## kill

Send signals to processes by PID.

```
kill 1234                   # send SIGTERM to PID 1234
kill -s KILL 1234           # send SIGKILL
kill -9 1234                # shorthand for SIGKILL
kill -l                     # list available signals
```

**Signals:** `HUP` (1), `INT` (2), `KILL` (9), `TERM` (15).

## Building

All utilities are built automatically by the build recipe:

```
gobake build                # builds dush + all utilities
gobake build-all            # cross-compile for all platforms
```

Binaries output to `build/`.
