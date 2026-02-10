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
	"github.com/josinSbazin/gf/internal/output"
	"github.com/spf13/cobra"
)

type commentOptions struct {
	repo    string
	body    string
	file    string
	line    int
	oldLine int
}

func newCommentCmd() *cobra.Command {
	opts := &commentOptions{}

	cmd := &cobra.Command{
		Use:   "comment <id>",
		Short: "Add a comment to a merge request",
		Long: `Add a comment to a merge request.

Without --body flag, opens an interactive prompt for the comment text.
Use --file and --line to add inline comments on specific lines.`,
		Example: `  # Add general comment
  gf mr comment 42 --body "Looks good to me!"

  # Add inline comment on a specific file and line
  gf mr comment 42 --body "This needs error handling" --file "main.go" --line 42

  # Add inline comment on removed line (old side of diff)
  gf mr comment 42 --body "Why was this removed?" --file "utils.go" --old-line 15

  # Pipe comment from stdin
  echo "LGTM" | gf mr comment 42 --body -`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runComment(opts, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")
	cmd.Flags().StringVarP(&opts.body, "body", "b", "", "Comment body (use - to read from stdin)")
	cmd.Flags().StringVarP(&opts.file, "file", "f", "", "File path for inline comment")
	cmd.Flags().IntVarP(&opts.line, "line", "l", 0, "New-side line number for inline comment")
	cmd.Flags().IntVar(&opts.oldLine, "old-line", 0, "Old-side line number for inline comment")

	return cmd
}

func runComment(opts *commentOptions, id int) error {
	// Validate inline comment flags
	if opts.file != "" && opts.line == 0 && opts.oldLine == 0 {
		return fmt.Errorf("--file requires --line or --old-line")
	}
	if (opts.line > 0 || opts.oldLine > 0) && opts.file == "" {
		return fmt.Errorf("--line/--old-line requires --file")
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

	// Get MR info first
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Get comment body
	body := opts.body
	if body == "-" {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		body = strings.Join(lines, "\n")
	} else if body == "" {
		// Interactive mode
		fmt.Printf("Adding comment to MR #%d: %s\n\n", mr.LocalID, mr.Title)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Comment: ")
		body, _ = reader.ReadString('\n')
		body = strings.TrimSpace(body)
	}

	if body == "" {
		return fmt.Errorf("comment body cannot be empty")
	}

	// Build discussion request
	req := &api.CreateDiscussionRequest{
		Message: body,
	}

	// Add inline comment fields if specified
	// GitFlic API requires all four fields (newLine, oldLine, newPath, oldPath)
	// and all line values must be > 0
	if opts.file != "" {
		req.NewPath = &opts.file
		req.OldPath = &opts.file
		newLine := opts.line
		oldLine := opts.oldLine
		// If only one side specified, mirror to the other (API requires both > 0)
		if newLine > 0 && oldLine == 0 {
			oldLine = newLine
		}
		if oldLine > 0 && newLine == 0 {
			newLine = oldLine
		}
		req.NewLine = &newLine
		req.OldLine = &oldLine
	}

	// Create discussion
	_, err = client.MergeRequests().CreateDiscussion(repo.Owner, repo.Name, id, req)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	if opts.file != "" {
		lineInfo := ""
		if opts.line > 0 {
			lineInfo = fmt.Sprintf(":%d", opts.line)
		} else if opts.oldLine > 0 {
			lineInfo = fmt.Sprintf(":%d (old)", opts.oldLine)
		}
		fmt.Printf("✓ Added inline comment to MR #%d on %s%s\n", mr.LocalID, opts.file, lineInfo)
	} else {
		fmt.Printf("✓ Added comment to MR #%d\n", mr.LocalID)
	}
	return nil
}

func newCommentsCmd() *cobra.Command {
	opts := &struct {
		repo string
	}{}

	cmd := &cobra.Command{
		Use:     "comments <id>",
		Aliases: []string{"discussions"},
		Short:   "List comments on a merge request",
		Long:    `List all comments and discussions on a merge request, grouped by file.`,
		Example: `  # List comments
  gf mr comments 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if err != nil {
				return fmt.Errorf("invalid merge request ID: %s", args[0])
			}
			return runComments(opts.repo, id)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "R", "", "Repository (owner/name)")

	return cmd
}

func runComments(repoFlag string, id int) error {
	// Get repository
	repo, err := git.ResolveRepo(repoFlag, config.DefaultHost())
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

	// Get MR info first
	mr, err := client.MergeRequests().Get(repo.Owner, repo.Name, id)
	if err != nil {
		if api.IsNotFound(err) {
			return fmt.Errorf("merge request #%d not found in %s", id, repo.FullName())
		}
		return fmt.Errorf("failed to get merge request: %w", err)
	}

	// Try threaded format first
	threads, err := client.MergeRequests().ListDiscussionThreads(repo.Owner, repo.Name, id)
	if err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	if len(threads) == 0 {
		fmt.Printf("No comments on MR #%d: %s\n", mr.LocalID, mr.Title)
		return nil
	}

	fmt.Printf("\nComments on MR #%d: %s\n", mr.LocalID, mr.Title)
	fmt.Println(strings.Repeat("─", 60))

	// Separate inline and general comments
	var inlineThreads []api.DiscussionThread
	var generalThreads []api.DiscussionThread

	for _, t := range threads {
		if t.RootNote.NewPath != nil && *t.RootNote.NewPath != "" {
			inlineThreads = append(inlineThreads, t)
		} else {
			generalThreads = append(generalThreads, t)
		}
	}

	// Print inline comments grouped by file
	if len(inlineThreads) > 0 {
		// Group by file
		fileGroups := make(map[string][]api.DiscussionThread)
		var fileOrder []string
		for _, t := range inlineThreads {
			path := *t.RootNote.NewPath
			if _, exists := fileGroups[path]; !exists {
				fileOrder = append(fileOrder, path)
			}
			fileGroups[path] = append(fileGroups[path], t)
		}

		fmt.Println("\n📁 Inline comments:")
		for _, path := range fileOrder {
			fmt.Printf("\n  %s\n", path)
			for _, t := range fileGroups[path] {
				printThreadedDiscussion(t, "    ")
			}
		}
	}

	// Print general comments
	if len(generalThreads) > 0 {
		fmt.Println("\n💬 General comments:")
		for _, t := range generalThreads {
			printThreadedDiscussion(t, "  ")
		}
	}

	fmt.Println()
	return nil
}

func printThreadedDiscussion(t api.DiscussionThread, indent string) {
	root := t.RootNote

	// Status badge
	status := ""
	if root.Resolved {
		status = " ✅"
	}

	// Line info
	lineInfo := ""
	if root.NewLine != nil && *root.NewLine > 0 {
		lineInfo = fmt.Sprintf(":%d", *root.NewLine)
	} else if root.OldLine != nil && *root.OldLine > 0 {
		lineInfo = fmt.Sprintf(":%d (old)", *root.OldLine)
	}

	// UUID (short)
	shortUUID := root.UUID
	if len(shortUUID) > 8 {
		shortUUID = shortUUID[:8]
	}

	authorName := root.Author.Username
	if authorName == "" {
		authorName = root.Author.FullName
	}

	fmt.Printf("%s%s @%s • %s%s  [%s]\n",
		indent, lineInfo, authorName,
		output.FormatRelativeTime(root.CreatedAt), status, shortUUID)
	fmt.Printf("%s%s\n", indent, root.Message)

	// Print replies
	for _, reply := range t.Replies {
		replyAuthor := reply.Author.Username
		if replyAuthor == "" {
			replyAuthor = reply.Author.FullName
		}
		fmt.Printf("%s  └─ @%s • %s\n", indent, replyAuthor, output.FormatRelativeTime(reply.CreatedAt))
		fmt.Printf("%s     %s\n", indent, reply.Message)
	}
}
