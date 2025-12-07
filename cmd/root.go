// ABOUTME: Root command setup for gh-context CLI extension
// ABOUTME: Configures Cobra command tree and shared output helpers

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-context",
	Short: "A kubectx-style context switcher for GitHub CLI",
	Long: `gh-context: Manage multiple GitHub accounts and hosts easily.

Switch between personal, work, and enterprise GitHub accounts
without manually managing authentication each time.

Contexts are stored in: ~/.config/gh/contexts/ (or %APPDATA%\gh\contexts on Windows)`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(currentCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(bindCmd)
	rootCmd.AddCommand(unbindCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(shellHookCmd)
	rootCmd.AddCommand(authStatusCmd)
}

// Output helpers that match the bash script style

// printErr prints an error message with ✗ prefix.
func printErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "✗ "+format+"\n", a...)
}

// printInfo prints an informational message with • prefix.
func printInfo(format string, a ...interface{}) {
	fmt.Printf("• "+format+"\n", a...)
}

// printOk prints a success message with ✓ prefix.
func printOk(format string, a ...interface{}) {
	fmt.Printf("✓ "+format+"\n", a...)
}

// printPlain prints a message without prefix.
func printPlain(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
