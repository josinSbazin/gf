package repo

import (
	"github.com/spf13/cobra"
)

// NewCmdRepo returns the repo command group
func NewCmdRepo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Work with repositories",
		Long:  `View and manage repositories.`,
	}

	cmd.AddCommand(newViewCmd())

	return cmd
}
