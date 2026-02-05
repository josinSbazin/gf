package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type loginOptions struct {
	hostname string
	token    string
	stdin    bool
}

func newLoginCmd() *cobra.Command {
	opts := &loginOptions{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a GitFlic host",
		Long: `Authenticate with a GitFlic host.

The token can be obtained from GitFlic settings:
  Profile Settings → API Tokens → Create`,
		Example: `  # Interactive login
  gf auth login

  # Login with token from argument
  gf auth login --token gf_xxxxxxxxxxxx

  # Login to a self-hosted instance
  gf auth login --hostname git.company.com

  # Login from CI (read token from stdin)
  echo $GF_TOKEN | gf auth login --stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.hostname, "hostname", "h", config.DefaultHost(), "GitFlic hostname")
	cmd.Flags().StringVarP(&opts.token, "token", "t", "", "Access token")
	cmd.Flags().BoolVar(&opts.stdin, "stdin", false, "Read token from stdin")

	return cmd
}

func runLogin(opts *loginOptions) error {
	var token string
	var err error

	if opts.stdin {
		// Read token from stdin
		reader := bufio.NewReader(os.Stdin)
		token, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read token from stdin: %w", err)
		}
		token = strings.TrimSpace(token)
	} else if opts.token != "" {
		token = opts.token
	} else {
		// Interactive mode
		fmt.Printf("GitFlic hostname: %s\n", opts.hostname)
		fmt.Print("Paste your access token: ")

		// Read password without echo
		tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		fmt.Println() // newline after hidden input
		token = string(tokenBytes)
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Verify token by calling /user/me
	baseURL := config.BaseURL(opts.hostname)
	client := api.NewClient(baseURL, token)

	user, err := client.Users().Me()
	if err != nil {
		if api.IsUnauthorized(err) {
			return fmt.Errorf("invalid token")
		}
		return fmt.Errorf("failed to verify token: %w", err)
	}

	// Save to config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.SetHost(opts.hostname, &config.Host{
		Token:    token,
		User:     user.Alias,
		Protocol: "https",
	})
	cfg.ActiveHost = opts.hostname

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Logged in as %s to %s\n", user.Alias, opts.hostname)
	return nil
}
