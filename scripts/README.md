This directory contains automation scripts for tasks such as building, testing, or deployment of the `dush` project.

## `build.go` Script

The `build.go` script is used to compile the `dush` executable for various operating systems and architectures, including support for different build configurations.

### Usage

To run the build script, use the `go run` command followed by the script path and optional arguments:

`go run scripts/build.go [target_os] [target_arch] [build_type]`

#### Arguments:

*   **`target_os`** (optional):
    *   If omitted, defaults to the operating system where the script is being run (`runtime.GOOS`).
    *   Can be a specific OS (e.g., `linux`, `windows`, `darwin`).
    *   Use `all` to build for all predefined supported operating systems and architectures.
*   **`target_arch`** (optional):
    *   If omitted, defaults to the architecture where the script is being run (`runtime.GOARCH`).
    *   Can be a specific architecture (e.g., `amd64`, `arm64`).
*   **`build_type`** (optional):
    *   `normal` (default): Compiles the application without any special build tags. This is the standard production build.
    *   `test`: Compiles the application with the `test` build tag enabled. This build will use the test-specific configuration (e.g., `internal/config/config.piml`).

### Examples:

1.  **Default build (current OS/Arch, normal type):**
    `go run scripts/build.go`
    *Output:* `build/dush-<current_os>-<current_arch>` (e.g., `dush-windows-amd64.exe`)

2.  **Test build for current OS/Arch:**
    `go run scripts/build.go test`
    or `go run scripts/build.go <current_os> <current_arch> test`
    *Output:* `build/dush-<current_os>-<current_arch>-test` (e.g., `dush-windows-amd64-test.exe`)

3.  **Normal build for a specific OS/Arch:**
    `go run scripts/build.go linux amd64`
    *Output:* `build/dush-linux-amd64`

4.  **Test build for a specific OS/Arch:**
    `go run scripts/build.go windows arm64 test`
    *Output:* `build/dush-windows-arm64-test.exe`

5.  **Build all supported targets (normal and test for each):**
    `go run scripts/build.go all`
    *This will generate multiple executables in the `build/` directory, for example:*
    *   `dush-linux-amd64`
    *   `dush-linux-amd64-test`
    *   `dush-windows-amd64.exe`
    *   `dush-windows-amd64-test.exe`
    *   ... and so on for all defined targets.

The compiled executables will be placed in the `build/` directory.