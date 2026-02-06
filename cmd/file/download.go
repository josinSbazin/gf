package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type downloadOptions struct {
	repo   string
	ref    string
	output string
}

func newDownloadCmd() *cobra.Command {
	opts := &downloadOptions{}

	cmd := &cobra.Command{
		Use:   "download <path>",
		Short: "Download a file",
		Long:  `Download a file from the repository.`,
		Example: `  # Download file to current directory
  gf file download README.md

  # Download to specific path
  gf file download src/main.go --output ./downloads/main.go

  # Download from specific branch
  gf file download config.json --ref develop`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Branch or tag name (default: default branch)")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Output path (default: filename)")

	return cmd
}

func runDownload(opts *downloadOptions, path string) error {
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

	// Determine output path with path traversal protection
	outputPath := opts.output
	if outputPath == "" {
		outputPath = filepath.Base(path)
	}

	// Security: sanitize output path to prevent path traversal
	outputPath = sanitizeOutputPath(outputPath)
	if outputPath == "" {
		return fmt.Errorf("invalid output path")
	}

	// Download file
	body, err := client.Files().Download(repo.Owner, repo.Name, ref, path)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer body.Close()

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	written, err := io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Downloaded %s (%s)\n", outputPath, formatSize(written))
	return nil
}

// sanitizeOutputPath prevents path traversal attacks by cleaning the path
// and ensuring it doesn't escape the current directory (unless absolute path given by user)
func sanitizeOutputPath(path string) string {
	// Clean the path to resolve any . or .. components
	cleaned := filepath.Clean(path)

	// If user specified an absolute path, allow it (user's explicit choice)
	if filepath.IsAbs(path) {
		return cleaned
	}

	// For relative paths, ensure we don't escape current directory
	// Check if cleaned path tries to go up (starts with ..)
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return ""
	}

	// Additional check: ensure the path doesn't contain .. anywhere
	// This catches cases like "foo/../../../etc/passwd"
	if strings.Contains(cleaned, "..") {
		return ""
	}

	return cleaned
}
