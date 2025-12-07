// ABOUTME: New command for gh-context - creates a new saved context
// ABOUTME: Supports --from-current to capture current session or explicit parameters

package cmd

import (
	"fmt"
	"os"

	"github.com/pmorgan/gh-context/internal/auth"
	"github.com/pmorgan/gh-context/internal/config"
	"github.com/pmorgan/gh-context/internal/ssh"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new context",
	Long: `Create a new context from the current session or with explicit parameters.

For SSH transport, the SSH key is required. When using --from-current, it will
detect the currently active SSH key from your ~/.ssh/config file.

Examples:
  gh context new --from-current --name work
  gh context new --from-current --name personal --ssh-key ~/.ssh/id_personal
  gh context new --hostname github.com --user myuser --ssh-key ~/.ssh/id_mykey --name mycontext`,
	RunE: runNew,
}

var (
	newName        string
	newFromCurrent bool
	newHostname    string
	newUser        string
	newTransport   string
	newSSHKey      string
)

func init() {
	newCmd.Flags().StringVar(&newName, "name", "", "Context name (required)")
	newCmd.Flags().BoolVar(&newFromCurrent, "from-current", false, "Create context from current gh session")
	newCmd.Flags().StringVar(&newHostname, "hostname", "", "GitHub hostname (default: github.com)")
	newCmd.Flags().StringVar(&newUser, "user", "", "GitHub username")
	newCmd.Flags().StringVar(&newTransport, "transport", "ssh", "Transport protocol (ssh or https)")
	newCmd.Flags().StringVar(&newSSHKey, "ssh-key", "", "Path to SSH key (e.g., ~/.ssh/id_personal)")

	newCmd.MarkFlagRequired("name")
}

func runNew(cmd *cobra.Command, args []string) error {
	// Validate context name
	if err := config.ValidateName(newName); err != nil {
		return err
	}

	// Check if context already exists
	exists, err := config.Exists(newName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("context '%s' already exists", newName)
	}

	var hostname, user, sshKey string

	if newFromCurrent {
		// Get from current session
		hostname = newHostname
		if hostname == "" {
			hostname = os.Getenv("GH_HOST")
		}
		if hostname == "" {
			hostname = "github.com"
		}

		// Get current user from API
		currentUser, authErr := auth.GetCurrentUserFromSession(hostname)
		if authErr != nil {
			printErr("Could not detect current user on '%s'", hostname)
			printInfo("Make sure you're logged in: gh auth login --hostname %s", hostname)
			return fmt.Errorf("authentication required")
		}

		user = newUser
		if user == "" {
			user = currentUser
		}

		// Get SSH key - from flag or detect from current config
		sshKey = newSSHKey
		if sshKey == "" && newTransport == "ssh" {
			// Try to detect from SSH config
			sshCfg, err := ssh.ParseConfig("")
			if err == nil {
				activeKey := sshCfg.GetActiveIdentityFile(hostname)
				if activeKey != "" {
					sshKey = activeKey
					printInfo("Detected SSH key from config: %s", sshKey)
				}
			}
		}
	} else {
		// Explicit parameters required
		if newHostname == "" || newUser == "" {
			return fmt.Errorf("provide either --from-current or both --hostname and --user")
		}
		hostname = newHostname
		user = newUser
		sshKey = newSSHKey
	}

	// Validate transport
	switch newTransport {
	case "ssh", "https":
		// Valid
	default:
		return fmt.Errorf("transport must be 'ssh' or 'https', got: %s", newTransport)
	}

	// For SSH transport, require SSH key
	if newTransport == "ssh" && sshKey == "" {
		printErr("SSH key is required for SSH transport")
		printInfo("Provide --ssh-key PATH or ensure your ~/.ssh/config has an active IdentityFile for %s", hostname)
		return fmt.Errorf("SSH key required")
	}

	// Validate SSH key exists if provided
	if sshKey != "" && !ssh.KeyExists(sshKey) {
		printErr("SSH key file not found: %s", ssh.ExpandPath(sshKey))
		return fmt.Errorf("SSH key not found")
	}

	// Create and save context
	ctx := &config.Context{
		Name:      newName,
		Hostname:  hostname,
		User:      user,
		Transport: newTransport,
		SSHKey:    sshKey,
	}

	if err := ctx.Save(); err != nil {
		return err
	}

	sshInfo := ""
	if sshKey != "" {
		sshInfo = fmt.Sprintf(", key=%s", sshKey)
	}

	printOk("Created context '%s' → %s@%s (%s%s)", newName, user, hostname, newTransport, sshInfo)
	return nil
}
