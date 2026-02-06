package auth

import (
	"fmt"

	"github.com/josinSbazin/gf/internal/config"
	"github.com/spf13/cobra"
)

type logoutOptions struct {
	hostname string
	all      bool
}

func newLogoutCmd() *cobra.Command {
	opts := &logoutOptions{}

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of a GitFlic host",
		Long: `Remove authentication for a GitFlic host.

This will remove the stored token from the configuration file.`,
		Example: `  # Logout from default host
  gf auth logout

  # Logout from specific host
  gf auth logout --hostname git.company.com

  # Logout from all hosts
  gf auth logout --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.hostname, "hostname", "H", "", "GitFlic hostname to logout from")
	cmd.Flags().BoolVar(&opts.all, "all", false, "Logout from all hosts")

	return cmd
}

func runLogout(opts *logoutOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if opts.all {
		// Remove all hosts
		count := len(cfg.Hosts)
		if count == 0 {
			fmt.Println("Not logged in to any hosts")
			return nil
		}

		cfg.Hosts = make(map[string]*config.Host)
		cfg.ActiveHost = config.DefaultHost()

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Logged out from %d host(s)\n", count)
		return nil
	}

	// Determine which host to logout from
	hostname := opts.hostname
	if hostname == "" {
		hostname = cfg.ActiveHost
	}
	if hostname == "" {
		hostname = config.DefaultHost()
	}

	// Check if logged in
	host := cfg.GetHost(hostname)
	if host == nil {
		return fmt.Errorf("not logged in to %s", hostname)
	}

	// Remove the host
	delete(cfg.Hosts, hostname)

	// If this was the active host, set a new active host
	if cfg.ActiveHost == hostname {
		cfg.ActiveHost = config.DefaultHost()
		// If we have other hosts, use one of them
		for h := range cfg.Hosts {
			cfg.ActiveHost = h
			break
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Logged out of %s\n", hostname)
	return nil
}
