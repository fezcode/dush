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

func performBuild(targetOS, targetArch, buildType, appName string) error {
	// Create output file name based on app name, target OS, Arch, and Build Type
	outputFileName := fmt.Sprintf("%s-%s-%s", appName, targetOS, targetArch)
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

	fmt.Printf("Building %s executable for %s/%s (Build Type: %s)...\n", appName, targetOS, targetArch, strings.ToUpper(buildType))

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

	// Set version
	version := "0.1.0"

	buildArgs := []string{"build"}
	// Use buildinfo paths only for dush, others might not have it
	ldflags := ""
	if appName == "dush" {
		ldflags = fmt.Sprintf("-s -w -X 'dush/cmd/dush/buildinfo.Version=%s' -X 'dush/cmd/dush/buildinfo.Commit=%s' -X 'dush/cmd/dush/buildinfo.BuildDate=%s'", version, commit, buildDate)
		if buildType == "test" {
			ldflags += " -X 'dush/cmd/dush/buildinfo.isTestBuild=true'"
		}
	} else {
		ldflags = "-s -w"
	}

	if ldflags != "" {
		buildArgs = append(buildArgs, "-ldflags="+ldflags)
	}

	buildArgs = append(buildArgs, "-o", outputFilePath, fmt.Sprintf("./cmd/%s", appName))

	cmd := exec.Command("go", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "."

	// Set GOOS and GOARCH for cross-compilation in the command's environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", targetOS))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", targetArch))
	cmd.Env = append(cmd.Env, "CGO_ENABLED=0")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error building %s for %s/%s (Build Type: %s): %v", appName, targetOS, targetArch, strings.ToUpper(buildType), err)
	}

	fmt.Printf("%s executable built successfully: %s\n", appName, outputFilePath)
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
		}
		argIndex++
	}

	// Find all apps in cmd/
	apps := []string{}
	entries, _ := os.ReadDir("cmd")
	for _, e := range entries {
		if e.IsDir() && e.Name() != "commands" && e.Name() != "buildinfo" {
			// Check if main.go exists
			if _, err := os.Stat(fmt.Sprintf("cmd/%s/main.go", e.Name())); err == nil {
				apps = append(apps, e.Name())
			}
		}
	}

	if targetOS == "all" {
		fmt.Println("Building all targets for all apps...")
		for _, bt := range []string{"normal", "test"} {
			for _, t := range targets {
				for _, appName := range apps {
					if err := performBuild(t.OS, t.Arch, bt, appName); err != nil {
						fmt.Printf("Failed to build %s %s/%s %s: %v\n", appName, t.OS, t.Arch, bt, err)
						os.Exit(1)
					}
				}
			}
		}
	} else {
		for _, appName := range apps {
			if err := performBuild(targetOS, targetArch, buildType, appName); err != nil {
				fmt.Printf("Build failed for %s.\n", appName)
				os.Exit(1)
			}
		}
	}

	fmt.Println("All requested builds completed.")
}
