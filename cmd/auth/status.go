package auth

import (
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/spf13/cobra"
)

type statusOptions struct {
	hostname string
}

func newStatusCmd() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "View authentication status",
		Long:  `View authentication status for configured GitFlic hosts.`,
		Example: `  # Check status for all hosts
  gf auth status

  # Check status for specific host
  gf auth status --hostname git.company.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.hostname, "hostname", "H", "", "Check specific hostname")

	return cmd
}

func runStatus(opts *statusOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Hosts) == 0 {
		fmt.Println("Not logged in to any GitFlic hosts.")
		fmt.Println("Run 'gf auth login' to authenticate.")
		return nil
	}

	// If specific hostname requested, check only that
	if opts.hostname != "" {
		host := cfg.GetHost(opts.hostname)
		if host == nil {
			fmt.Printf("%s\n  ✗ Not logged in\n", opts.hostname)
			return nil
		}
		return checkHost(opts.hostname, host)
	}

	// Check all configured hosts
	for hostname, host := range cfg.Hosts {
		if err := checkHost(hostname, host); err != nil {
			fmt.Printf("  ✗ Error: %s\n", err)
		}
		fmt.Println()
	}

	return nil
}

func checkHost(hostname string, host *config.Host) error {
	fmt.Println(hostname)

	// Try to verify token
	baseURL := config.BaseURL(hostname)
	client := api.NewClient(baseURL, host.Token)

	user, err := client.Users().Me()
	if err != nil {
		if api.IsUnauthorized(err) {
			fmt.Println("  ✗ Token expired or invalid")
			return nil
		}
		fmt.Printf("  ✗ Could not verify: %s\n", err)
		return nil
	}

	fmt.Printf("  ✓ Logged in as %s\n", user.Username)

	// Show masked token
	token := host.Token
	if len(token) > 8 {
		fmt.Printf("  ✓ Token: %s...%s\n", token[:4], token[len(token)-4:])
	}

	return nil
}
