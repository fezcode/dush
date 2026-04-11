# Configuration

All config lives in `~/.dush/`:

| File | Loaded | Purpose |
|---|---|---|
| `env` | Always (interactive + scripts) | Environment setup, pub vars, prompt config |
| `is` | Interactive only | Aliases, greeting, interactive helpers |
| `history` | Interactive only | Command history (auto-managed) |

## Loading Order

1. `~/.dush/env` -- always loaded (interactive + scripts)
2. `~/.dush/is` -- interactive only

## CLI Flags

```
dush                          Interactive mode
dush script.dush              Run a script (non-interactive)
dush -v / --version           Print version
dush -h / --help              Print help
```

### Path Overrides

```
dush --env <path>             Override env file (default: ~/.dush/env)
dush --is <path>              Override is file (default: ~/.dush/is)
dush --history <path>         Override history file (default: ~/.dush/history)
```

Useful for testing or isolated sessions:
```
dush --env ./test/env --is /dev/null --history /tmp/h examples/showcase.dush
```

## Non-interactive Mode

Running `dush script.dush` only sources `~/.dush/env`, skipping `~/.dush/is`. Scripts run fast without interactive setup.

## Prompt

The prompt is customizable via `@PROMPT_LINE` in `~/.dush/env` using `{token}` syntax.

### Default Prompt

```
@PROMPT_LINE = '{fg:cyan}{user}{reset}@{fg:green}{dir}{reset} {fg:yellow}${reset} '
```

### Content Tokens

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
| `{os}` | Operating system |
| `{newline}` | Line break |

### Color and Style Tokens

| Token | Effect |
|---|---|
| `{fg:red}` | Foreground color (named) |
| `{bg:blue}` | Background color (named) |
| `{fg:#ff5500}` | Foreground color (hex) |
| `{bg:#1a1a2e}` | Background color (hex) |
| `{reset}` | Reset all colors/styles |
| `{bold}` | Bold text |
| `{dim}` | Dim text |
| `{italic}` | Italic text |
| `{underline}` | Underlined text |

**Named colors:** `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` -- plus `bright_*` variants.

### Continuation Prompt

```
@CONTINUATION_PROMPT = '{fg:yellow}...{reset} '
```

### Example Prompts

```
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
```

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
