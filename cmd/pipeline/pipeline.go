package pipeline

import (
	"github.com/spf13/cobra"
)

// NewCmdPipeline returns the pipeline command group
func NewCmdPipeline() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pipeline",
		Aliases: []string{"ci", "pl"},
		Short:   "Work with CI/CD pipelines",
		Long:    `View and manage CI/CD pipelines.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newWatchCmd())
	cmd.AddCommand(newRetryCmd())
	cmd.AddCommand(newCancelCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newJobCmd())

	return cmd
}
