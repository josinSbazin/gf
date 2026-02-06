package release

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type createOptions struct {
	repo         string
	title        string
	notes        string
	notesFile    string
	isDraft      bool
	isPrerelease bool
}

func newCreateCmd() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create <tag>",
		Short: "Create a release",
		Long: `Create a new release for an existing tag.

Note: The tag must already exist in the repository. Push your tag first with:
  git tag v1.0.0
  git push origin v1.0.0`,
		Example: `  # Create a release for tag v1.0.0
  gf release create v1.0.0

  # Create with title and notes
  gf release create v1.0.0 --title "Version 1.0" --notes "First release"

  # Create a draft release
  gf release create v1.0.0 --draft

  # Create a pre-release
  gf release create v1.0.0 --prerelease`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Release title (defaults to tag name)")
	cmd.Flags().StringVarP(&opts.notes, "notes", "n", "", "Release notes")
	cmd.Flags().StringVarP(&opts.notesFile, "notes-file", "F", "", "Read release notes from file")
	cmd.Flags().BoolVarP(&opts.isDraft, "draft", "d", false, "Save as draft")
	cmd.Flags().BoolVarP(&opts.isPrerelease, "prerelease", "p", false, "Mark as pre-release")

	return cmd
}

func runCreate(opts *createOptions, tagName string) error {
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

	// Determine title
	title := opts.title
	if title == "" {
		title = tagName
	}

	// Get release notes
	notes := opts.notes
	if opts.notesFile != "" {
		data, err := os.ReadFile(opts.notesFile)
		if err != nil {
			return fmt.Errorf("failed to read notes file: %w", err)
		}
		notes = string(data)
	}

	// If no notes provided, open editor or prompt
	if notes == "" && !opts.isDraft {
		fmt.Print("Release notes (press Enter twice to finish):\n")
		notes = readMultiline()
	}

	// Create release request
	req := &api.CreateReleaseRequest{
		Title:        title,
		Description:  notes,
		TagName:      tagName,
		IsDraft:      opts.isDraft,
		IsPrerelease: opts.isPrerelease,
	}

	// Create release
	release, err := client.Releases().Create(repo.Owner, repo.Name, req)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("tag '%s' not found. Push the tag first:\n  git tag %s\n  git push origin %s", tagName, tagName, tagName)
		}
		return fmt.Errorf("failed to create release: %w", err)
	}

	releaseType := "Release"
	if release.IsDraft {
		releaseType = "Draft release"
	} else if release.IsPrerelease {
		releaseType = "Pre-release"
	}

	fmt.Printf("%s '%s' created for tag %s\n", releaseType, release.Title, release.TagName)
	// GitFlic uses release UUID in web URLs, not tag name
	fmt.Printf("https://%s/project/%s/%s/release/%s\n",
		repo.Host, repo.Owner, repo.Name, release.ID)

	return nil
}

func readMultiline() string {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	emptyCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			emptyCount++
			if emptyCount >= 2 {
				break
			}
		} else {
			emptyCount = 0
		}
		lines = append(lines, line)
	}

	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}
