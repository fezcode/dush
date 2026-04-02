# Dush Language Specification (v1 & Roadmap)

This document defines the syntax and semantics of the Dush scripting language.

## 1. Core Philosophy
Dush is a strictly typed, C-style scripting language that treats shell commands as first-class citizens. It differentiates strictly between **Code** (variables, logic) and **Commands** (shell execution).

## 2. Basic Syntax
- **Blocks:** Defined by curly braces `{}`.
- **Statement Terminators:** New-lines. Semicolons are optional.
- **Comments:** Single-line comments start with `//`.

## 3. Variables and Data Types
- **Declaration:** `let x = value`
- **Assignment:** `x = new_value`
- **Access:** **Strict Rule**: You MUST use `var(name)` to access the value of a data variable in expressions or command arguments.
  - Correct: `echo var(x)`
  - Incorrect: `echo x` (This executes the command `x`)
- **Types:** `int`, `string`, `bool`.
- **Planned Types:** `list`, `map`, `float`.

## 4. Control Flow
- **Conditionals:**
  - `if (condition) { ... }`
  - `else { ... }`
- **Loops:**
  - **Conditional:** `loop (condition) { ... }`
  - **Iterator:** `loop (item : collection) { ... }` (Currently supports iterating string characters)

## 5. Procedures
- **Declaration:** `proc name(arg1, arg2) { ... }`
- **Call:** `name(arg1, arg2)`
- **Note:** Procedure names are resolvable directly, unlike data variables.

## 6. Shell Integration
- **Bare Words:** Any identifier that is not a keyword or defined procedure is treated as a **Command**.
  - `git status` -> Runs `git` with argument `status`.
  - `x` -> Runs command `x`.
- **Arguments:**
  - Strings: `"hello world"`
  - Variables: `name` (Automatically resolved if it exists locally or in OS environment, otherwise treated as string literal `"name"`)
  - Function Calls: `len("str")`
  - Flags: `-la` (Parsed as strings)
- **Command Chaining:**
  - `&&`: Run next if previous succeeded (Exit Code 0).
  - `||`: Run next if previous failed.
- **Exit Status:**
  - The special variable `LAST_STATUS` holds the exit code of the last run command.

## 7. Built-in Functions & Commands
- **Commands:** `ls`, `cd`, `echo`, `export`, `source`, `.`.
- **Functions:**
  - `len(str)`: Returns length of string.
  - `var(name)`: Returns value of variable.

---

## 8. Startup Profile
When `dush` is started as an interactive shell, it automatically looks for and executes a `~/.dushis` (Dush Interactive Shell) file in the user's home directory. You can use this file to configure environment variables, create aliases, and define procedures.

---

## 8. Roadmap / Planned Features (Draft)
The following features are designed but not yet implemented in v1.

- **Pipes & Redirection:**
  - `ls | grep "go"`
  - `echo "hello" > file.txt`
- **Output Capture:**
  - `let files = save(ls)` (Capture command output to variable)
- **Globbing:**
  - `ls *.go` (Wildcard expansion)
- **Scoped Blocks (`with`):**
  - `with (NODE_ENV="production") { ... }`
- **String Library:**
  - `split(str, sep)`, `replace(str, old, new)`, `to_upper(str)`
- **File Library:**
  - `exists(path)`, `is_dir(path)``exists(path)`, `is_dir(path)`