package mr

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type editOptions struct {
	repo  string
	title string
	body  string
}

func newEditCmd() *cobra.Command {
	opts := &editOptions{}

	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a merge request",
		Long:  `Edit the title or description of a merge request.`,
		Example: `  # Edit MR interactively
  gf mr edit 42

  # Edit title
  gf mr edit 42 --title "New title"

  # Edit description
  gf mr edit 42 --body "New description"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runEdit(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "New title")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "New description")

	return cmd
}

func runEdit(opts *editOptions, id int) error {
	// Get repository
	repo, err := git.ResolveRepo(opts.repo, config.DefaultHost())
	if err != nil {
		return fmt.Errorf("could not determine repository: %w\nUse --repo owner/name to specify", err)
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

	// Get current MR info
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Interactive mode if no flags provided
	if opts.title == "" && opts.body == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Editing MR #%d: %s\n\n", mr.LocalID, mr.Title)

		fmt.Printf("Title [%s]: ", mr.Title)
		newTitle, _ := reader.ReadString('\n')
		newTitle = strings.TrimSpace(newTitle)
		if newTitle != "" {
			opts.title = newTitle
		}

		fmt.Printf("Description [press Enter to keep current]: ")
		newBody, _ := reader.ReadString('\n')
		newBody = strings.TrimSpace(newBody)
		if newBody != "" {
			opts.body = newBody
		}

		if opts.title == "" && opts.body == "" {
			fmt.Println("No changes made.")
			return nil
		}
	}

	// Build update request
	req := &api.UpdateMRRequest{}
	if opts.title != "" {
		req.Title = opts.title
	}
	if opts.body != "" {
		req.Description = opts.body
	}

	// Update MR
	_, err = client.MergeRequests().Update(repo.Owner, repo.Name, id, req)
	if err != nil {
		return fmt.Errorf("failed to update merge request: %w", err)
	}

	fmt.Printf("✓ Updated merge request #%d\n", mr.LocalID)
	return nil
}
