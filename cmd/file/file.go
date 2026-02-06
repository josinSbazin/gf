package file

import (
	"github.com/spf13/cobra"
)

// NewCmdFile returns the file command group
func NewCmdFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "file",
		Aliases: []string{"f", "blob"},
		Short:   "Work with repository files",
		Long:    `List, view, and download repository files.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}
