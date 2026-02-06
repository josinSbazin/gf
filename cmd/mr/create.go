package mr

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

// validMRBranchRegex validates branch names for MR creation
var validMRBranchRegex = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9/_\.]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

// validateMRBranch checks if a branch name is valid for MR creation
func validateMRBranch(name, branchType string) error {
	if name == "" {
		return fmt.Errorf("%s branch cannot be empty", branchType)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("%s branch cannot start with '-'", branchType)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("%s branch cannot contain '..'", branchType)
	}
	if !validMRBranchRegex.MatchString(name) {
		return fmt.Errorf("invalid %s branch name: %s", branchType, name)
	}
	return nil
}

type createOptions struct {
	title        string
	body         string
	target       string
	source       string
	draft        bool
	deleteBranch bool
	repo         string
	web          bool
	quiet        bool
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
	cmd.Flags().BoolVarP(&opts.quiet, "quiet", "q", false, "Output only the MR number")

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

	// Validate branch names
	if err := validateMRBranch(opts.source, "source"); err != nil {
		return err
	}
	if err := validateMRBranch(opts.target, "target"); err != nil {
		return err
	}

	client := api.NewClient(config.BaseURL(cfg.ActiveHost), token)

	// Get project info to get UUID
	project, err := client.Projects().Get(repo.Owner, repo.Name)
	if err != nil {
		return fmt.Errorf("failed to get project info: %w", err)
	}

	// Create merge request
	mr, err := client.MergeRequests().Create(repo.Owner, repo.Name, &api.CreateMRRequest{
		Title:        opts.title,
		Description:  opts.body,
		SourceBranch: api.BranchRef{ID: opts.source},
		TargetBranch: api.BranchRef{ID: opts.target},
		SourceProject: api.ProjectRef{ID: project.ID},
		TargetProject: api.ProjectRef{ID: project.ID},
		IsDraft:            opts.draft,
		RemoveSourceBranch: opts.deleteBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to create merge request: %w", err)
	}

	// Quiet mode - output only ID
	if opts.quiet {
		fmt.Printf("%d\n", mr.LocalID)
		return nil
	}

	// Regular output with draft indication
	if opts.draft {
		fmt.Printf("\n✓ Created draft merge request #%d\n", mr.LocalID)
	} else {
		fmt.Printf("\n✓ Created merge request #%d\n", mr.LocalID)
	}

	url := fmt.Sprintf("https://%s/project/%s/%s/merge-request/%d",
		repo.Host, repo.Owner, repo.Name, mr.LocalID)
	fmt.Println(url)

	if opts.web {
		return browser.Open(url)
	}

	return nil
}
