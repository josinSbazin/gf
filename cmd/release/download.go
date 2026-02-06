package release

import (
	"encoding/json"
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
	output string
	all    bool
	list   bool
}

func newDownloadCmd() *cobra.Command {
	opts := &downloadOptions{}

	cmd := &cobra.Command{
		Use:   "download <tag> [asset-name]",
		Short: "Download release assets",
		Long: `Download assets from a release.

Without an asset name, lists available assets.
Use --all to download all assets.`,
		Example: `  # List available assets
  gf release download v1.0.0 --list

  # Download specific asset
  gf release download v1.0.0 myapp-linux-amd64

  # Download all assets
  gf release download v1.0.0 --all

  # Download to specific path
  gf release download v1.0.0 myapp.zip --output ./downloads/`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			assetName := ""
			if len(args) > 1 {
				assetName = args[1]
			}
			return runDownload(opts, args[0], assetName)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Output path (file or directory)")
	cmd.Flags().BoolVarP(&opts.all, "all", "a", false, "Download all assets")
	cmd.Flags().BoolVarP(&opts.list, "list", "l", false, "List available assets")

	return cmd
}

func runDownload(opts *downloadOptions, tagName, assetName string) error {
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

	// Get assets
	assets, err := client.Releases().ListAssets(repo.Owner, repo.Name, tagName)
	if err != nil {
		return fmt.Errorf("failed to list assets: %w", err)
	}

	if len(assets) == 0 {
		fmt.Printf("No assets in release %s\n", tagName)
		return nil
	}

	// List mode
	if opts.list || (assetName == "" && !opts.all) {
		fmt.Printf("\nAssets in release %s:\n\n", tagName)
		data, _ := json.MarshalIndent(assets, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Download all assets
	if opts.all {
		outputDir := opts.output
		if outputDir == "" {
			outputDir = "."
		}

		// Create output directory if needed
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		for _, asset := range assets {
			// Security: sanitize asset name to prevent path traversal
			safeName := sanitizeAssetName(asset.Name)
			if safeName == "" {
				fmt.Printf("⚠ Skipping asset with invalid name: %q\n", asset.Name)
				continue
			}
			outputPath := filepath.Join(outputDir, safeName)
			if err := downloadAsset(client, repo.Owner, repo.Name, tagName, asset.Name, outputPath); err != nil {
				return err
			}
		}
		fmt.Printf("\n✓ Downloaded %d assets to %s\n", len(assets), outputDir)
		return nil
	}

	// Download specific asset
	// Security: sanitize asset name to prevent path traversal
	safeAssetName := sanitizeAssetName(assetName)
	if safeAssetName == "" {
		return fmt.Errorf("invalid asset name: %q", assetName)
	}

	outputPath := opts.output
	if outputPath == "" {
		outputPath = safeAssetName
	} else {
		// If output is a directory, use asset name as filename
		if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
			outputPath = filepath.Join(outputPath, safeAssetName)
		}
	}

	return downloadAsset(client, repo.Owner, repo.Name, tagName, assetName, outputPath)
}

// sanitizeAssetName prevents path traversal attacks by ensuring
// the asset name is a simple filename without directory components
func sanitizeAssetName(name string) string {
	// Get just the base name, removing any directory components
	base := filepath.Base(filepath.Clean(name))

	// Reject empty, current dir, or parent dir references
	if base == "" || base == "." || base == ".." {
		return ""
	}

	// Reject if it still contains path separators or traversal patterns
	if strings.ContainsAny(base, "/\\") || strings.Contains(base, "..") {
		return ""
	}

	return base
}

func downloadAsset(client *api.Client, owner, project, tagName, assetName, outputPath string) error {
	fmt.Printf("Downloading %s...\n", assetName)

	body, _, err := client.Releases().DownloadAsset(owner, project, tagName, assetName)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("asset %q not found in release %s", assetName, tagName)
		}
		return fmt.Errorf("failed to download asset: %w", err)
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
