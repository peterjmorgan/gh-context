// ABOUTME: Current command for gh-context - shows active context and repo binding
// ABOUTME: Displays current context details and any repository-specific bindings

package cmd

import (
	"fmt"

	"github.com/peterjmorgan/gh-context/internal/config"
	"github.com/peterjmorgan/gh-context/internal/git"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active context and repo-bound context",
	Long:  `Display the currently active context and any repository-specific context binding.`,
	RunE:  runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
	active, err := config.GetActive()
	if err != nil {
		return err
	}

	if active == "" {
		printPlain("No active context")
	} else {
		ctx, loadErr := config.Load(active)
		if loadErr != nil {
			printErr("Active context '%s' points to missing file", active)
			printInfo("Run 'gh context list' to see available contexts")
			return nil
		}

		sshInfo := ""
		if ctx.SSHKey != "" {
			sshInfo = fmt.Sprintf(", key=%s", ctx.SSHKey)
		}

		printPlain("Active: %s (%s@%s, %s%s)", ctx.Name, ctx.User, ctx.Hostname, ctx.Transport, sshInfo)
	}

	// Check for repo binding
	root, err := git.RepoRoot()
	if err != nil {
		return err
	}

	if root != "" {
		binding, err := git.GetBinding()
		if err != nil {
			return err
		}
		if binding != "" {
			bindingPath, _ := git.BindingPath()
			printPlain("Repo-bound: %s (in %s)", binding, bindingPath)
		}
	}

	return nil
}
