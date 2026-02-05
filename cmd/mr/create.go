package mr

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type createOptions struct {
	title        string
	body         string
	target       string
	source       string
	draft        bool
	deleteBranch bool
	repo         string
	web          bool
}

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a merge request",
		Long:  `Create a new merge request.`,
		Example: `  # Interactive create
  gf mr create

  # Create with title
  gf mr create --title "Add new feature"

  # Create with all options
  gf mr create --title "Fix bug" --body "Description" --target main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Title of the merge request")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Description of the merge request")
	cmd.Flags().StringVarP(&opts.target, "target", "T", "", "Target branch (default: main)")
	cmd.Flags().StringVarP(&opts.source, "source", "S", "", "Source branch (default: current branch)")
	cmd.Flags().BoolVar(&opts.draft, "draft", false, "Create as draft")
	cmd.Flags().BoolVarP(&opts.deleteBranch, "delete-branch", "d", false, "Delete source branch after merge")
	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.web, "web", "w", false, "Open in browser after creating")

	return cmd
}

func runCreate(opts *createOptions) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w", err)
	}

	// Get source branch
	if opts.source == "" {
		opts.source, err = git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}
	}

	// Get target branch
	if opts.target == "" {
		opts.target, err = git.DefaultBranch()
		if err != nil {
			opts.target = "main" // fallback
		}
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

	// Interactive mode if title not provided
	if opts.title == "" {
		fmt.Printf("Creating merge request for %s into %s in %s\n\n",
			opts.source, opts.target, repo.FullName())

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Title: ")
		opts.title, _ = reader.ReadString('\n')
		opts.title = strings.TrimSpace(opts.title)

		if opts.title == "" {
			return fmt.Errorf("title is required")
		}

		fmt.Print("Description (optional, press Enter to skip): ")
		opts.body, _ = reader.ReadString('\n')
		opts.body = strings.TrimSpace(opts.body)
	}

	client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)

	// Create merge request
	mr, err := client.MergeRequests().Create(repo.Owner, repo.Name, &api.CreateMRRequest{
		Title:              opts.title,
		Description:        opts.body,
		SourceBranch:       opts.source,
		TargetBranch:       opts.target,
		IsDraft:            opts.draft,
		RemoveSourceBranch: opts.deleteBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to create merge request: %w", err)
	}

	fmt.Printf("\n✓ Created merge request #%d\n", mr.LocalID)

	url := fmt.Sprintf("https://%s/project/%s/%s/merge-request/%d",
		repo.Host, repo.Owner, repo.Name, mr.LocalID)
	fmt.Println(url)

	if opts.web {
		return browser.Open(url)
	}

	return nil
}
