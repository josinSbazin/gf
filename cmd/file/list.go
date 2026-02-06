package file

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/config"
	"github.com/josinSbazin/gf/internal/git"
	"github.com/spf13/cobra"
)

type listOptions struct {
	repo string
	ref  string
	json bool
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: "List files in a directory",
		Long:  `List files and directories at the given path.`,
		Example: `  # List root directory
  gf file list

  # List specific directory
  gf file list src/

  # List on specific branch
  gf file list src/ --ref develop

  # Output as JSON
  gf file list --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runList(opts, path)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVar(&opts.ref, "ref", "", "Branch or tag name (default: default branch)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output as JSON")

	return cmd
}

func runList(opts *listOptions, path string) error {
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

	// List files
	entries, err := client.Files().List(repo.Owner, repo.Name, ref, path)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("path not found: %s", path)
		}
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("(empty directory)")
		return nil
	}

	// JSON output
	if opts.json {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Print tree-like output
	for _, e := range entries {
		size := ""
		if e.Size > 0 {
			size = formatSize(e.Size)
		}

		// Use Name() method to get just the filename from path
		name := e.Name()
		fmt.Printf("📄 %-40s %s\n", name, size)
	}

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

// isTextFile checks if the file might be text based on extension
func isTextFile(path string) bool {
	textExtensions := []string{
		".txt", ".md", ".json", ".yaml", ".yml", ".xml", ".html", ".css", ".js", ".ts",
		".go", ".py", ".rb", ".java", ".c", ".cpp", ".h", ".rs", ".sh", ".bash",
		".gitignore", ".env", ".toml", ".ini", ".cfg", ".conf", ".log",
	}
	for _, ext := range textExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}
	return false
}
