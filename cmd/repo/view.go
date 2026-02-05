package repo

import (
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	web  bool
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view [owner/name]",
		Short: "View a repository",
		Long:  `View details of a repository.`,
		Example: `  # View current repository
  gf repo view

  # View specific repository
  gf repo view owner/name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.repo = args[0]
			}
			return runView(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser")

	return cmd
}

func runView(opts *viewOptions) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w\nUse 'gf repo view owner/name' to specify", err)
	}

	// Load config and create client
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	token, err := cfg.Token()
	if err != nil {
		return fmt.Errorf("not authenticated. Run 'gf auth login' first")
	}

	client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)

	// Fetch project
	project, err := client.Projects().Get(repo.Owner, repo.Name)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("repository %s/%s not found", repo.Owner, repo.Name)
		}
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Print details
	fmt.Printf("\n%s/%s\n", repo.Owner, project.Alias)

	if project.Description != "" {
		fmt.Println(project.Description)
	}

	fmt.Println()

	visibility := "Public"
	if project.IsPrivate {
		visibility = "Private"
	}
	fmt.Printf("Visibility: %s\n", visibility)

	if project.StarsCount > 0 {
		fmt.Printf("Stars:      %d\n", project.StarsCount)
	}
	if project.ForksCount > 0 {
		fmt.Printf("Forks:      %d\n", project.ForksCount)
	}
	if project.Language != "" {
		fmt.Printf("Language:   %s\n", project.Language)
	}

	fmt.Println()

	// URLs
	httpsURL := fmt.Sprintf("https://%s/project/%s/%s.git", repo.Host, repo.Owner, project.Alias)
	sshURL := fmt.Sprintf("git@%s:%s/%s.git", repo.Host, repo.Owner, project.Alias)

	fmt.Printf("Clone URL: %s\n", httpsURL)
	fmt.Printf("SSH URL:   %s\n", sshURL)

	fmt.Println()

	webURL := fmt.Sprintf("https://%s/project/%s/%s", repo.Host, repo.Owner, project.Alias)

	if opts.web {
		fmt.Printf("Opening %s in browser...\n", webURL)
		return browser.Open(webURL)
	}

	fmt.Printf("View in browser: %s\n", webURL)

	return nil
}
