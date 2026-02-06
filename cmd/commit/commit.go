package commit

import (
	"github.com/spf13/cobra"
)

// NewCmdCommit returns the commit command group
func NewCmdCommit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "commit",
		Aliases: []string{"c"},
		Short:   "Work with commits",
		Long:    `List and view repository commits.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newDiffCmd())

	return cmd
}
