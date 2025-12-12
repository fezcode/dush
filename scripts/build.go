package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	// Determine target OS and Arch
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	// Check environment variables for cross-compilation targets
	if os.Getenv("GOOS") != "" {
		targetOS = os.Getenv("GOOS")
	}
	if os.Getenv("GOARCH") != "" {
		targetArch = os.Getenv("GOARCH")
	}

	// Override with command-line arguments if provided
	if len(os.Args) > 1 {
		targetOS = os.Args[1]
		if len(os.Args) > 2 {
			targetArch = os.Args[2]
		}
	}

	// Create output file name based on target OS and Arch
	outputFileName := fmt.Sprintf("dush-%s-%s", targetOS, targetArch)
	if targetOS == "windows" {
		outputFileName += ".exe"
	}

	// Ensure the build directory exists
	buildDir := "build"
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		err := os.Mkdir(buildDir, 0755)
		if err != nil {
			fmt.Printf("Error creating build directory %s: %v\n", buildDir, err)
			os.Exit(1)
		}
	}

	outputFilePath := fmt.Sprintf("%s/%s", buildDir, outputFileName)

	fmt.Printf("Building dush executable for %s/%s...\n", targetOS, targetArch)

	cmd := exec.Command("go", "build", "-o", outputFilePath, "./cmd/dush")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "." // Changed from ".."

	// Set GOOS and GOARCH for cross-compilation in the command's environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", targetOS))
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOARCH=%s", targetArch))

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error building dush for %s/%s: %v\n", targetOS, targetArch, err)
		os.Exit(1)
	}

	fmt.Printf("dush executable built successfully: %s\n", outputFilePath)
	fmt.Println("To run this executable, navigate to the 'build' directory and execute it.")
}
