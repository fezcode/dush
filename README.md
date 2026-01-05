# Dush - A Go Terminal Shell Project

This project aims to develop a custom terminal shell written in Go.

## Project Goal
The primary goal of Dush is to create a functional and extensible command-line interface (CLI) that provides a user-friendly experience for interacting with the operating system. It will serve as a learning platform for understanding shell mechanics, Go programming, and system interactions.

## Features
- [x] **Command Execution**: Execute external programs and commands.
- [x] **Built-in Commands**: Implement essential shell built-in commands (e.g., `cd`, `exit`, `pwd`).
- [ ] **Input/Output Redirection**: Support basic I/O redirection (`<`, `>`, `>>`).
- [ ] **Piping**: Allow chaining commands with pipes (`|`).
- [ ] **Environment Variables**: Manage and access environment variables.
- [x] **Command History**: Basic command history for easy recall.
- [ ] **Customizable Prompt**: A dynamic and informative shell prompt.

## Getting Started

### Prerequisites
- Go (version 1.16 or higher) installed on your system.

### Setup
1.  **Clone the repository (if applicable):**
    ```bash
    git clone [repository-url]
    cd dush
    ```
2.  **Initialize Go module (if not already done):**
    ```bash
    go mod init dush
    ```
3.  **Build the shell:**
    ```bash
    go build -o dush ./cmd/dush
    ```
4.  **Run the shell:**
    ```bash
    ./dush
    ```

## Codebase Structure
A typical Go terminal shell project, incorporating best practices for CLI applications, could be organized as follows:

```
dush/
├── cmd/
│   └── dush/
│       ├── main.go         // Main entry point
│       └── commands/       // Defines commands, e.g., root.go, subcommands.go
├── internal/
│   ├── app/                // Core business logic (e.g., command execution, process management)
│   ├── parser/             // Handles parsing user input into abstract syntax trees or commands
│   ├── evaluator/          // Interprets and executes parsed commands
│   ├── builtins/           // Implements shell built-in commands (cd, exit, etc.)
│   ├── repl/               // Read-Eval-Print Loop logic
│   ├── config/             // Configuration loading and management
│   └── util/               // General utility functions specific to this application
├── pkg/                    // Reusable libraries/packages meant for external use (if applicable)
├── scripts/                // Build, deploy, or helper scripts
├── test/                   // Test files for the project
├── go.mod                  // Go module file
├── go.sum                  // Go module checksum file
├── LICENSE
└── README.md
```

*   **`cmd/`**: Contains `main` packages for executables. Each executable typically has its own subdirectory. `main.go` defines the root command and other subcommands, possibly leveraging a library like `Cobra`.
*   **`internal/`**: Private application code not intended for import by other projects.
    *   `app/`: Core application logic, often separated into services or components related to shell operations.
    *   `parser/`: Logic for parsing the command line input.
    *   `evaluator/`: Logic for evaluating and executing parsed commands.
    *   `builtins/`: Implementations of commands internal to the shell.
    *   `repl/`: Handles the Read-Eval-Print Loop for interactive sessions.
    *   `config/`: Handles application configuration (loading, parsing, validating), potentially using `Viper`.
    *   `util/`: Small, general-purpose utility functions specific to this application.
*   **`pkg/`**: Public libraries that other applications might import. Use only if you intend to publish reusable packages.
*   **`scripts/`**: Automation scripts for building, testing, or deployment.

## Contribution
Contributions are welcome! Please follow the standard Go contribution guidelines.

