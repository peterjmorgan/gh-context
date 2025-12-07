// ABOUTME: Delete command for gh-context - removes a saved context
// ABOUTME: Clears active pointer if deleted context was active

package cmd

import (
	"github.com/peterjmorgan/gh-context/internal/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "remove"},
	Short:   "Remove a saved context",
	Long:    `Delete a saved context. Clears the active pointer if the deleted context was active.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if we need to clear active pointer
	active, _ := config.GetActive()
	willClearActive := active == name

	if err := config.Delete(name); err != nil {
		return err
	}

	if willClearActive {
		printInfo("Cleared active context pointer")
	}

	printOk("Deleted context '%s'", name)
	return nil
}
