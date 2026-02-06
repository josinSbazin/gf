package release

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type uploadOptions struct {
	repo string
	name string
}

func newUploadCmd() *cobra.Command {
	opts := &uploadOptions{}

	cmd := &cobra.Command{
		Use:   "upload <tag> <file>",
		Short: "Upload an asset to a release",
		Long: `Upload a file as an asset to an existing release.

The file will be available for download on the release page.`,
		Example: `  # Upload a binary
  gf release upload v1.0.0 ./dist/myapp-linux-amd64

  # Upload with custom name
  gf release upload v1.0.0 ./build/app.zip --name myapp-v1.0.0.zip`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpload(opts, args[0], args[1])
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.name, "name", "n", "", "Asset name (default: file name)")

	return cmd
}

func runUpload(opts *uploadOptions, tagName, filePath string) error {
	// Validate file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot upload directory: %s", filePath)
	}

	// Get file name
	fileName := opts.name
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}

	// Security: validate asset name to prevent path traversal on server
	fileName = sanitizeAssetName(fileName)
	if fileName == "" {
		return fmt.Errorf("invalid asset name")
	}

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

	// Check if release exists
	_, err = client.Releases().Get(repo.Owner, repo.Name, tagName)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("release %q not found in %s", tagName, repo.FullName())
		}
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload file
	fmt.Printf("Uploading %s (%s)...\n", fileName, formatSize(fileInfo.Size()))
	asset, err := client.Releases().UploadAsset(repo.Owner, repo.Name, tagName, fileName, file)
	if err != nil {
		if api.IsForbidden(err) {
			return fmt.Errorf("permission denied: you don't have access to upload assets in %s", repo.FullName())
		}
		return fmt.Errorf("failed to upload asset: %w", err)
	}

	fmt.Printf("✓ Uploaded %q to release %s\n", asset.Name, tagName)
	return nil
}

// formatSize formats a file size in human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
