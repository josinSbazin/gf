package auth

import (
	"github.com/spf13/cobra"
)

// NewCmdAuth returns the auth command group
func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with GitFlic",
		Long:  `Manage authentication state for GitFlic.`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newStatusCmd())

	return cmd
}
