package issue

import (
	"github.com/spf13/cobra"
)

// NewCmdIssue returns the issue command group
func NewCmdIssue() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "issue",
		Aliases: []string{"i"},
		Short:   "Work with issues",
		Long:    `Create, view, and manage project issues.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newCloseCmd())
	cmd.AddCommand(newReopenCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newCommentCmd())
	cmd.AddCommand(newCommentsCmd())

	return cmd
}
