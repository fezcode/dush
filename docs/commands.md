# Built-in Commands

All built-in commands support `-h` or `--help` for usage information.

## Navigation

| Command | Description |
|---|---|
| `cd [path]` | Change directory. No args = home directory |
| `pwd` | Print working directory |
| `pushdir <dir>` | Push current directory onto stack, cd to dir |
| `popdir` | Pop directory from stack and cd to it |
| `dirs [-c]` | Show directory stack. `-c` clears it |

## Files & Directories

| Command | Description |
|---|---|
| `ls [opts] [path]` | List directory contents (icons, colors, grid/table) |
| `mkdir [-pv] <dir>...` | Create directories. `-p` creates parents |
| `mkfile [-pvs] <file>...` | Create files. `-s SIZE` sets file size |
| `rm [-rfv] <target>...` | Remove files/directories. `-r` recursive, `-f` force |
| `touch [-acm] <file>...` | Update timestamps or create files. `-c` no-create |

## Output

| Command | Description |
|---|---|
| `echo [-n] [args...]` | Print arguments. `-n` omits trailing newline |
| `clear` | Clear the terminal screen |

## Shell State

| Command | Description |
|---|---|
| `alias [name[=value]]` | Define or display aliases |
| `unalias <name>` | Remove an alias |
| `history` | Show command history with line numbers |
| `help` | List all built-in commands |
| `version` | Show dush version, commit, and build date |
| `source <file>` / `. <file>` | Execute a file in the current environment |
| `read [@var...]` | Read a line from stdin into variables (default: `@REPLY`) |
| `unset @var...` | Remove variables from the environment |
| `whatis <name>...` | Show how a name resolves (alias, builtin, or external path) |
| `trap <cmd> <signal>...` | Run a command when a signal is received |

## Job Control

| Command | Description |
|---|---|
| `command &` | Run any command in the background |
| `jobs` | List all background jobs with status and PID |
| `fg <id>` | Wait for a background job to finish |
| `wait [id...]` | Wait for background jobs. No args = wait for all |
| `killjob <id>` | Kill a running background job |
| `disown <id>` | Remove a job from the shell's job table |
| `cleanjobs` | Remove finished jobs from the list |

## Other

| Command | Description |
|---|---|
| `sleep <duration>` | Pause for seconds (`5`) or Go duration (`1.5s`, `10m`) |
| `export @VAR = value` | *Deprecated* -- use `pub @VAR = value` instead |

---

## `ls` in Detail

Modern `ls` with icons, colors, and grid/table layout.

**Flags:**

| Flag | Long | Description |
|---|---|---|
| `-l` | `--long` | Long table format with permissions, size, owner, created/modified times |
| `-a` | `--all` | Show hidden (dot) files |
| `-1` | `--oneline` | One entry per line |
| `-r` | `--reverse` | Reverse sort order |
| `-S` | `--sort=size` | Sort by file size |
| `-t` | `--sort=time` | Sort by modification time |
| `-X` | `--sort=ext` | Sort by file extension |
| | `--no-header` | Hide the table header |
| | `--no-icons` | Hide file type icons |
| `-h` | `--help` | Show help |

**Examples:**
```
ls                  Grid view with icons and colors
ls -la              Long format with hidden files
ls -lS              Sorted by size
ls -lt              Sorted by time
ls -r               Reverse sort
```

## `rm` in Detail

**Flags:**

| Flag | Long | Description |
|---|---|---|
| `-f` | `--force` | Ignore nonexistent files, never prompt |
| `-r` / `-R` | `--recursive` | Remove directories and contents recursively |
| `-v` | `--verbose` | Explain what is being done |

## `mkdir` in Detail

**Flags:**

| Flag | Long | Description |
|---|---|---|
| `-p` | `--parents` | No error if existing, make parents as needed |
| `-v` | `--verbose` | Print a message for each created directory |

## `mkfile` in Detail

**Flags:**

| Flag | Long | Description |
|---|---|---|
| `-p` | `--parents` | Make parent directories as needed |
| `-s` | `--size=SIZE` | Set file size (e.g., `10k`, `5m`, `2g`, `512b`) |
| `-v` | `--verbose` | Print a message for each created file |

## `touch` in Detail

**Flags:**

| Flag | Long | Description |
|---|---|---|
| `-a` | | Change only access time |
| `-m` | | Change only modification time |
| `-c` | `--no-create` | Do not create files |

## `trap` in Detail

```
trap 'echo bye' EXIT        Run on shell exit
trap '' INT                  Remove INT trap
trap -l                     List active traps
```

**Supported signals:** `INT`, `TERM`, `HUP`, `EXIT`, `ERR`

## `read` in Detail

```
read @name                   Read a line into @name
read @first @last            Split by words: first word -> @first, rest -> @last
read                         Store entire line in @REPLY
```

## `whatis` in Detail

```
whatis ls                    ls is a shell builtin
whatis git                   git is /usr/bin/git
whatis ll                    ll is aliased to 'ls -la'
```
