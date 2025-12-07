// ABOUTME: Use command for gh-context - switches to a saved context
// ABOUTME: Sets active context, activates SSH key, and tests authentication

package cmd

import (
	"fmt"

	"github.com/peterjmorgan/gh-context/internal/auth"
	"github.com/peterjmorgan/gh-context/internal/config"
	"github.com/peterjmorgan/gh-context/internal/ssh"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to context (updates SSH config and gh auth)",
	Long: `Switch to a saved context. This will:
1. Set the active context
2. Update ~/.ssh/config to use the correct SSH key
3. Switch gh CLI authentication to the correct user

If authentication is not configured, provides instructions to set it up.`,
	Args: cobra.ExactArgs(1),
	RunE: runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load context to verify it exists
	ctx, loadErr := config.Load(name)
	if loadErr != nil {
		// Context not found - show available contexts
		contexts, listErr := config.List()
		if listErr == nil && len(contexts) > 0 {
			printErr("Context '%s' not found", name)
			printInfo("Available contexts: %v", contexts)
		}
		return loadErr
	}

	// Set context immediately (fast by default)
	if err := config.SetActive(name); err != nil {
		return err
	}

	printOk("Switched to context '%s' (%s@%s)", name, ctx.User, ctx.Hostname)

	// Activate SSH key if configured
	if ctx.SSHKey != "" && ctx.Transport == "ssh" {
		printInfo("Activating SSH key: %s", ctx.SSHKey)

		sshCfg, err := ssh.ParseConfig("")
		if err != nil {
			printErr("Failed to read SSH config: %v", err)
		} else {
			err = sshCfg.ActivateKey(ctx.Hostname, ctx.SSHKey)
			if err != nil {
				printErr("Failed to activate SSH key: %v", err)
				printInfo("You may need to manually update your ~/.ssh/config")
			} else {
				if err := sshCfg.Save(); err != nil {
					printErr("Failed to save SSH config: %v", err)
				} else {
					printOk("SSH config updated (backup saved to ~/.ssh/config.bak)")
				}
			}
		}
	}

	// Test if authentication works
	printInfo("Testing authentication...")
	authenticated, testErr := auth.TestAuth(ctx.Hostname, ctx.User)
	if testErr == nil && authenticated {
		printOk("Authentication verified")
		return nil
	}

	// Authentication failed - prompt user to fix it
	printErr("Authentication required for %s@%s", ctx.User, ctx.Hostname)
	fmt.Println()
	printInfo("Your context has been set, but authentication is needed.")
	printInfo("Please authenticate and your context will work automatically:")
	fmt.Println()
	printInfo("  gh auth login --hostname %s --username %s --scopes repo,read:org", ctx.Hostname, ctx.User)
	fmt.Println()
	printInfo("After authentication, all gh commands will use the correct account.")

	return nil
}
