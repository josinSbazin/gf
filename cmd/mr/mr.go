package mr

import (
	"github.com/spf13/cobra"
)

// NewCmdMR returns the mr command group
func NewCmdMR() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mr",
		Aliases: []string{"merge-request"},
		Short:   "Work with merge requests",
		Long:    `Create, view, and manage merge requests.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newMergeCmd())

	return cmd
}
