// ABOUTME: List command for gh-context - shows all saved contexts
// ABOUTME: Displays context names with active indicator and configuration details

package cmd

import (
	"fmt"

	"github.com/pmorgan/gh-context/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all contexts with active indicator",
	Long:    `List all saved contexts, showing which one is currently active.`,
	RunE:    runList,
}

func runList(cmd *cobra.Command, args []string) error {
	contexts, err := config.ListContexts()
	if err != nil {
		return err
	}

	if len(contexts) == 0 {
		printInfo("No contexts found. Create one with: gh context new --from-current --name <name>")
		return nil
	}

	active, err := config.GetActive()
	if err != nil {
		return err
	}

	printPlain("Available contexts:")
	for _, ctx := range contexts {
		indicator := ""
		if ctx.Name == active {
			indicator = " *"
		}

		sshInfo := ""
		if ctx.SSHHostAlias != "" {
			sshInfo = fmt.Sprintf(", ssh_host=%s", ctx.SSHHostAlias)
		}

		fmt.Printf("  %s%s\t(%s@%s, %s%s)\n",
			ctx.Name, indicator, ctx.User, ctx.Hostname, ctx.Transport, sshInfo)
	}

	if active != "" {
		fmt.Println()
		printPlain("* = active context")
	}

	return nil
}
