package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dush",
	Short: "Dush is a custom terminal shell.",
	Long: `Dush is a custom terminal shell written in Go.
It aims to provide a functional and extensible command-line interface.`, // Corrected: Removed unnecessary escaping of newline
	// Uncomment the following line if your bare application has an action
	// Run: func(cmd *cobra.Command, args []string) { fmt.Println("Hello from Dush!") },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be available to all subcommands.
	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dush.yaml)")

	// Cobra also supports local flags, which will only run when this command
	// is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
