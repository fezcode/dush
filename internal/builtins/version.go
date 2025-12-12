package builtins

import (
	"context" // New import
	"fmt"
	"io"

	"dush/cmd/dush/buildinfo"
)

// VersionCommand represents the 'version' built-in command.
type VersionCommand struct{}

// Execute prints the version information of the application.
func (c *VersionCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("version command does not accept arguments")
	}

	fmt.Fprintf(out, "Dush Version: %s\n", buildinfo.Version)
	fmt.Fprintf(out, "Commit: %s\n", buildinfo.Commit)
	fmt.Fprintf(out, "Build Date: %s\n", buildinfo.BuildDate)
	if buildinfo.IsTestBuild() {
		fmt.Fprintln(out, "This is a TEST build.")
	}
	return nil
}

func init() {
	RegisterBuiltin("version", &VersionCommand{})
}
