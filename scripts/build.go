package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Target represents an OS/Arch combination
type Target struct {
	OS   string
	Arch string
}

// Global list of common build targets
var targets = []Target{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"}, // For Apple Silicon
}

func performBuild(targetOS, targetArch, buildType string) error {
	// Create output file name based on target OS, Arch, and Build Type
	outputFileName := fmt.Sprintf("dush-%s-%s", targetOS, targetArch)
	if buildType == "test" {
		outputFileName += "-test"
	}
	if targetOS == "windows" {
		outputFileName += ".exe"
	}

	// Ensure the build directory exists
	buildDir := "build"
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		err := os.Mkdir(buildDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating build directory %s: %v", buildDir, err)
		}
	}

	outputFilePath := fmt.Sprintf("%s/%s", buildDir, outputFileName)

	fmt.Printf("Building dush executable for %s/%s (Build Type: %s)...\n", targetOS, targetArch, strings.ToUpper(buildType))

	// Get current git commit hash
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitOutput, err := commitCmd.Output()
	if err != nil {
		fmt.Printf("Warning: Could not get git commit hash: %v. Using 'unknown'.\n", err)
		commitOutput = []byte("unknown")
	}
	commit := strings.TrimSpace(string(commitOutput))

	// Get current build date
	buildDate := time.Now().Format(time.RFC3339)

	// Set version (can be dynamic, but for now, hardcode)
	version := "0.1.0"

	buildArgs := []string{"build"}
	ldflags := fmt.Sprintf("-X 'dush/cmd/dush/buildinfo.Version=%s' -X 'dush/cmd/dush/buildinfo.Commit=%s' -X 'dush/cmd/dush/buildinfo.BuildDate=%s'", version, commit, buildDate)
	if buildType == "test" {
		ldflags += " -X 'dush/cmd/dush/buildinfo.isTestBuild=true'"
	}

	if ldflags != "" {
		buildArgs = append(buildArgs, "-ldflags="+ldflags)
	}

	buildArgs = append(buildArgs, "-o", outputFilePath, "./cmd/dush")

	cmd := exec.Command("go", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."

	// Set GOOS and GOARCH for cross-compilation in the command's environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", targetOS))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", targetArch))

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error building dush for %s/%s (Build Type: %s): %v", targetOS, targetArch, strings.ToUpper(buildType), err)
	}

	fmt.Printf("dush executable built successfully: %s\n", outputFilePath)
	return nil
}

func main() {
	// Determine target OS and Arch
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	buildType := "normal" // Default build type

	// Check environment variables for cross-compilation targets
	if os.Getenv("GOOS") != "" {
		targetOS = os.Getenv("GOOS")
	}
	if os.Getenv("GOARCH") != "" {
		targetArch = os.Getenv("GOARCH")
	}

	// Override with command-line arguments if provided
	// os.Args[0] is the program name itself
	argIndex := 1
	if len(os.Args) > argIndex {
		targetOS = os.Args[argIndex]
		argIndex++
	}
	if len(os.Args) > argIndex {
		targetArch = os.Args[argIndex]
		argIndex++
	}
	if len(os.Args) > argIndex {
		bt := strings.ToLower(os.Args[argIndex])
		if bt == "test" || bt == "normal" || bt == "all" {
			buildType = bt
		} else {
			fmt.Printf("Warning: Unknown build type '%s'. Using default 'normal'.\n", bt)
		}
		argIndex++
	}

	if targetOS == "all" {
		// Building for all targets
		fmt.Println("Building all targets...")
		for _, bt := range []string{"normal", "test"} {
			for _, t := range targets {
				if err := performBuild(t.OS, t.Arch, bt); err != nil {
					fmt.Printf("Failed to build %s/%s %s: %v\n", t.OS, t.Arch, bt, err)
					os.Exit(1)
				}
			}
		}
	} else {
		if err := performBuild(targetOS, targetArch, buildType); err != nil {
			fmt.Println("Build failed.")
			os.Exit(1)
		}
	}

	fmt.Println("All requested builds completed.")
	fmt.Println("To run an executable, navigate to the 'build' directory and execute it.")
}
