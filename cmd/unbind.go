// ABOUTME: Unbind command for gh-context - removes repo context binding
// ABOUTME: Deletes .ghcontext file from repository root

package cmd

import (
	"github.com/peterjmorgan/gh-context/internal/git"
	"github.com/spf13/cobra"
)

var unbindCmd = &cobra.Command{
	Use:   "unbind",
	Short: "Remove .ghcontext from repo root",
	Long:  `Remove the repository's context binding by deleting the .ghcontext file.`,
	Args:  cobra.NoArgs,
	RunE:  runUnbind,
}

func runUnbind(cmd *cobra.Command, args []string) error {
	// Verify we're in a git repo
	root, err := git.RepoRoot()
	if err != nil {
		return err
	}
	if root == "" {
		printErr("Not inside a Git repository")
		return nil
	}

	// Check if binding exists
	hasBinding, bindErr := git.HasBinding()
	if bindErr != nil {
		return bindErr
	}

	if !hasBinding {
		printInfo("No repo binding found")
		return nil
	}

	if removeErr := git.RemoveBinding(); removeErr != nil {
		return removeErr
	}

	printOk("Removed repo binding")
	return nil
}
