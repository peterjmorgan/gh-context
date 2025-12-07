// ABOUTME: Auth-status command for gh-context - shows authentication status
// ABOUTME: Displays auth state for all saved contexts with verification

package cmd

import (
	"fmt"

	"github.com/pmorgan/gh-context/internal/auth"
	"github.com/pmorgan/gh-context/internal/config"
	"github.com/pmorgan/gh-context/internal/ssh"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "auth-status",
	Short: "Display authentication status for all contexts",
	Long:  `Show the authentication status for all saved contexts, indicating which are ready to use.`,
	Args:  cobra.NoArgs,
	RunE:  runAuthStatus,
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	printPlain("Authentication status for all contexts:")
	fmt.Println()

	contexts, err := config.ListContexts()
	if err != nil {
		return err
	}

	if len(contexts) == 0 {
		printInfo("No contexts found")
		return nil
	}

	active, _ := config.GetActive()

	// Get current SSH config state
	sshCfg, _ := ssh.ParseConfig("")

	for _, ctx := range contexts {
		indicator := ""
		if ctx.Name == active {
			indicator = " *"
		}

		fmt.Printf("Context: %s%s\n", ctx.Name, indicator)
		fmt.Printf("  Host: %s\n", ctx.Hostname)
		fmt.Printf("  User: %s\n", ctx.User)
		fmt.Printf("  Transport: %s\n", ctx.Transport)

		// Show SSH key info
		if ctx.SSHKey != "" {
			keyExists := "❌"
			if ssh.KeyExists(ctx.SSHKey) {
				keyExists = "✅"
			}
			fmt.Printf("  SSH Key: %s %s\n", ctx.SSHKey, keyExists)

			// Check if this key is active in SSH config
			if sshCfg != nil {
				activeKey := sshCfg.GetActiveIdentityFile(ctx.Hostname)
				if activeKey != "" && ssh.ExpandPath(activeKey) == ssh.ExpandPath(ctx.SSHKey) {
					fmt.Printf("  SSH Active: ✅ (currently active in ~/.ssh/config)\n")
				} else {
					fmt.Printf("  SSH Active: ❌ (not active in ~/.ssh/config)\n")
				}
			}
		}

		// Check authentication status
		authIcon := "❌"
		if auth.IsUserLoggedIn(ctx.Hostname, ctx.User) {
			authIcon = "✅"
		}

		fmt.Printf("  GH Auth: %s\n", authIcon)

		// Show login command if not authenticated
		if authIcon == "❌" {
			fmt.Printf("  To fix: gh auth login --hostname %s --username %s --scopes repo,read:org\n",
				ctx.Hostname, ctx.User)
		}

		fmt.Println()
	}

	if active != "" {
		printPlain("* = active context")
	}

	return nil
}
