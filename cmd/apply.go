// ABOUTME: Apply command for gh-context - applies repo's bound context
// ABOUTME: Reads .ghcontext from repo root and switches to that context

package cmd

import (
	"github.com/peterjmorgan/gh-context/internal/git"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Read .ghcontext in this repo and switch to it",
	Long:  `Apply the context bound to the current repository by reading .ghcontext and switching.`,
	Args:  cobra.NoArgs,
	RunE:  runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
	// Verify we're in a git repo
	root, err := git.RepoRoot()
	if err != nil {
		return err
	}
	if root == "" {
		printErr("Not inside a Git repository")
		return nil
	}

	// Get binding
	binding, bindErr := git.GetBinding()
	if bindErr != nil {
		return bindErr
	}
	if binding == "" {
		printErr("No .ghcontext file found in repository")
		printInfo("Create one with: gh context bind <name>")
		return nil
	}

	// Use the bound context (reuse the use command logic)
	return runUse(cmd, []string{binding})
}
