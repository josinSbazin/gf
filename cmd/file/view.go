package file

import (
	"encoding/base64"
	"fmt"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type viewOptions struct {
	repo string
	ref  string
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}

	cmd := &cobra.Command{
		Use:   "view <path>",
		Short: "View file contents",
		Long:  `View the contents of a file in the repository.`,
		Example: `  # View file
  gf file view README.md

  # View file on specific branch
  gf file view src/main.go --ref develop`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Branch or tag name (default: default branch)")

	return cmd
}

func runView(opts *viewOptions, path string) error {
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

	// Get ref (default branch if not specified)
	ref := opts.ref
	if ref == "" {
		defaultBranch, err := client.Branches().GetDefault(repo.Owner, repo.Name)
		if err != nil {
			return fmt.Errorf("failed to get default branch: %w", err)
		}
		ref = defaultBranch.Name
	}

	// Get file content
	file, err := client.Files().Get(repo.Owner, repo.Name, ref, path)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Decode content if base64
	content := file.Content
	if file.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return fmt.Errorf("failed to decode file content: %w", err)
		}
		content = string(decoded)
	}

	fmt.Print(content)
	return nil
}
