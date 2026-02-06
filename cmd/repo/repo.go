package repo

import (
	"github.com/spf13/cobra"
)

// NewCmdRepo returns the repo command group
func NewCmdRepo() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repo",
		Aliases: []string{"r"},
		Short:   "Work with repositories",
		Long:    `View and manage repositories.`,
	}

	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newCloneCmd())

	return cmd
}
