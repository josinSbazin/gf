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
	cmd.AddCommand(newCloseCmd())
	cmd.AddCommand(newCheckoutCmd())
	cmd.AddCommand(newApproveCmd())
	cmd.AddCommand(newDiffCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newReopenCmd())
	cmd.AddCommand(newReadyCmd())
	cmd.AddCommand(newCommentCmd())
	cmd.AddCommand(newCommentsCmd())
	cmd.AddCommand(newReplyCmd())
	cmd.AddCommand(newResolveCmd())
	cmd.AddCommand(newReviewCmd())

	return cmd
}
