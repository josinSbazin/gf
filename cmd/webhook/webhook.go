package webhook

import (
	"github.com/spf13/cobra"
)

// NewCmdWebhook returns the webhook command group
func NewCmdWebhook() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook",
		Aliases: []string{"hook"},
		Short:   "Work with webhooks",
		Long:    `List, create, and manage repository webhooks.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newTestCmd())

	return cmd
}
