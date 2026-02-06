package branch

import (
	"github.com/spf13/cobra"
)

// NewCmdBranch returns the branch command group
func NewCmdBranch() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "branch",
		Aliases: []string{"br"},
		Short:   "Work with branches",
		Long:    `List, create, and delete repository branches.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
