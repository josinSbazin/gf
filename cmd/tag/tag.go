package tag

import (
	"github.com/spf13/cobra"
)

// NewCmdTag returns the tag command group
func NewCmdTag() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tag",
		Aliases: []string{"t"},
		Short:   "Work with tags",
		Long:    `List, create, and delete repository tags.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())

	return cmd
}
