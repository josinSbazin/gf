package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/josinSbazin/gf/internal/browser"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type browseOptions struct {
	repo     string
	branch   bool
	settings bool
	issues   bool
	mrs      bool
	mr       bool // Open specific MR by number
	pipeline bool
}

func newBrowseCmd() *cobra.Command {
	opts := &browseOptions{}

	cmd := &cobra.Command{
		Use:   "browse [<number>]",
		Short: "Open repository in browser",
		Long: `Open the repository, issue, merge request, or pipeline in the web browser.

Without arguments, opens the repository home page.
With a number, opens that issue or MR (use flags to specify which).`,
		Example: `  # Open repository in browser
  gf browse

  # Open current branch
  gf browse --branch

  # Open issue #42
  gf browse 42

  # Open MR #10
  gf browse --mr 10

  # Open settings
  gf browse --settings

  # Open issues list
  gf browse --issues

  # Open MRs list
  gf browse --mrs`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var number int
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid number: %s", args[0])
				}
				number = n
			}
			return runBrowse(opts, number)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().BoolVarP(&opts.branch, "branch", "b", false, "Open current branch")
	cmd.Flags().BoolVarP(&opts.settings, "settings", "s", false, "Open repository settings")
	cmd.Flags().BoolVar(&opts.issues, "issues", false, "Open issues list")
	cmd.Flags().BoolVar(&opts.mrs, "mrs", false, "Open merge requests list")
	cmd.Flags().BoolVarP(&opts.mr, "mr", "m", false, "Open merge request (with number)")
	cmd.Flags().BoolVarP(&opts.pipeline, "pipeline", "p", false, "Open pipelines list")

	return cmd
}

func runBrowse(opts *browseOptions, number int) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w\nUse --repo owner/name to specify", err)
	}

	// Build URL
	baseURL := fmt.Sprintf("https://%s/project/%s/%s", repo.Host, repo.Owner, repo.Name)
	var targetURL string

	switch {
	case opts.settings:
		targetURL = baseURL + "/settings"
	case opts.issues:
		targetURL = baseURL + "/issue"
	case opts.mrs:
		targetURL = baseURL + "/merge-request"
	case opts.pipeline:
		targetURL = baseURL + "/cicd/pipeline"
	case opts.branch:
		branch, err := git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not get current branch: %w", err)
		}
		targetURL = baseURL + "/branch/" + url.PathEscape(branch)
	case opts.mr && number > 0:
		targetURL = baseURL + "/merge-request/" + strconv.Itoa(number)
	case opts.mr && number == 0:
		// No number provided with --mr, open MR list (same as --mrs)
		targetURL = baseURL + "/merge-request"
	case number > 0:
		// Default to issue when number provided without --mr flag
		targetURL = baseURL + "/issue/" + strconv.Itoa(number)
	default:
		targetURL = baseURL
	}

	fmt.Printf("Opening %s\n", targetURL)
	return browser.Open(targetURL)
}
