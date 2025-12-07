// ABOUTME: Bind command for gh-context - associates a repo with a context
// ABOUTME: Creates .ghcontext file in repository root

package cmd

import (
	"github.com/peterjmorgan/gh-context/internal/config"
	"github.com/peterjmorgan/gh-context/internal/git"
	"github.com/spf13/cobra"
)

var bindCmd = &cobra.Command{
	Use:   "bind <name>",
	Short: "Write .ghcontext in repo root",
	Long: `Bind the current repository to a context by creating a .ghcontext file.
When using shell hooks, the context will be automatically applied when entering this repo.`,
	Args: cobra.ExactArgs(1),
	RunE: runBind,
}

func runBind(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Verify context exists
	exists, err := config.Exists(name)
	if err != nil {
		return err
	}
	if !exists {
		_, loadErr := config.Load(name) // This will return proper "not found" error
		return loadErr
	}

	// Verify we're in a git repo
	root, err := git.RepoRoot()
	if err != nil {
		return err
	}
	if root == "" {
		printErr("Not inside a Git repository")
		return nil
	}

	// Create binding
	if err := git.SetBinding(name); err != nil {
		return err
	}

	bindingPath, _ := git.BindingPath()
	printOk("Bound repo to context '%s' (%s)", name, bindingPath)
	printInfo("Add .ghcontext to .gitignore if you don't want to commit it")

	return nil
}
