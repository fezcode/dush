//go:build gobake
package bake_recipe

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fezcode/gobake"
)

func gitCommit() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func buildLdflags(version string) string {
	commit := gitCommit()
	date := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	pkg := "dush/cmd/dush/buildinfo"
	return fmt.Sprintf("-X main.Version=%s -X %s.Version=%s -X %s.Commit=%s -X %s.BuildDate=%s",
		version, pkg, version, pkg, commit, pkg, date)
}

func Run(bake *gobake.Engine) error {
	if err := bake.LoadRecipeInfo("recipe.piml"); err != nil {
		return err
	}

	bake.Task("build", "Builds the binary for the current platform", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s for %s/%s...", bake.Info.Name, bake.Info.Version, runtime.GOOS, runtime.GOARCH)

		err := ctx.Mkdir("build")
		if err != nil {
			return err
		}

		ldflags := buildLdflags(bake.Info.Version)

		// Find apps in cmd/
		entries, _ := os.ReadDir("cmd")
		var apps []string
		for _, e := range entries {
			if e.IsDir() && e.Name() != "commands" && e.Name() != "buildinfo" {
				if _, err := os.Stat(fmt.Sprintf("cmd/%s/main.go", e.Name())); err == nil {
					apps = append(apps, e.Name())
				}
			}
		}

		for _, appName := range apps {
			output := "build/" + appName
			if runtime.GOOS == "windows" {
				output += ".exe"
			}

			ctx.Env = []string{
				"CGO_ENABLED=0",
			}

			err := ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, "./cmd/"+appName)
			if err != nil {
				return err
			}
		}
		return nil
	})

	bake.Task("build-all", "Builds the binary for multiple platforms", func(ctx *gobake.Context) error {
		ctx.Log("Building %s v%s for all platforms...", bake.Info.Name, bake.Info.Version)

		targets := []struct {
			os   string
			arch string
		}{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"windows", "amd64"},
			{"windows", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
		}

		err := ctx.Mkdir("build")
		if err != nil {
			return err
		}

		ldflags := buildLdflags(bake.Info.Version)

		// Find apps in cmd/
		entries, _ := os.ReadDir("cmd")
		var apps []string
		for _, e := range entries {
			if e.IsDir() && e.Name() != "commands" && e.Name() != "buildinfo" {
				if _, err := os.Stat(fmt.Sprintf("cmd/%s/main.go", e.Name())); err == nil {
					apps = append(apps, e.Name())
				}
			}
		}

		for _, appName := range apps {
			for _, t := range targets {
				output := "build/" + appName + "-" + t.os + "-" + t.arch
				if t.os == "windows" {
					output += ".exe"
				}

				// We use manual go build to inject ldflags
				ctx.Env = []string{
					"CGO_ENABLED=0",
					"GOOS=" + t.os,
					"GOARCH=" + t.arch,
				}

				err := ctx.Run("go", "build", "-ldflags", ldflags, "-o", output, "./cmd/"+appName)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	bake.Task("clean", "Removes build artifacts", func(ctx *gobake.Context) error {
		return ctx.Remove("build")
	})

	return nil
}
