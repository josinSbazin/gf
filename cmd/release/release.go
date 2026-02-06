package release

import (
	"github.com/spf13/cobra"
)

// NewCmdRelease returns the release command group
func NewCmdRelease() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"rel"},
		Short:   "Work with releases",
		Long:    `Create, view, and manage releases.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newUploadCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}
