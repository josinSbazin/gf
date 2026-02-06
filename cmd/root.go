package cmd

import (
	"fmt"
	"os"

	"github.com/josinSbazin/gf/cmd/auth"
	"github.com/josinSbazin/gf/cmd/branch"
	"github.com/josinSbazin/gf/cmd/commit"
	"github.com/josinSbazin/gf/cmd/file"
	"github.com/josinSbazin/gf/cmd/issue"
	"github.com/josinSbazin/gf/cmd/mr"
	"github.com/josinSbazin/gf/cmd/pipeline"
	"github.com/josinSbazin/gf/cmd/release"
	"github.com/josinSbazin/gf/cmd/repo"
	"github.com/josinSbazin/gf/cmd/tag"
	"github.com/josinSbazin/gf/cmd/webhook"
	"github.com/josinSbazin/gf/internal/api"
	"github.com/josinSbazin/gf/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gf",
	Short: "GitFlic CLI - work with GitFlic from command line",
	Long: `gf is a CLI tool for GitFlic that brings merge requests,
pipelines, and more to your terminal.

Get started by running:
  gf auth login`,
	Version: version.Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// ExitError is used when a command wants to exit with specific code
		// (e.g., pipeline watch --exit-status). Don't print these as errors.
		if api.IsExitError(err) {
			os.Exit(api.GetExitCode(err))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceErrors = true
	rootCmd.AddCommand(newAPICmd())
	rootCmd.AddCommand(auth.NewCmdAuth())
	rootCmd.AddCommand(branch.NewCmdBranch())
	rootCmd.AddCommand(newBrowseCmd())
	rootCmd.AddCommand(commit.NewCmdCommit())
	rootCmd.AddCommand(file.NewCmdFile())
	rootCmd.AddCommand(issue.NewCmdIssue())
	rootCmd.AddCommand(mr.NewCmdMR())
	rootCmd.AddCommand(pipeline.NewCmdPipeline())
	rootCmd.AddCommand(release.NewCmdRelease())
	rootCmd.AddCommand(repo.NewCmdRepo())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(tag.NewCmdTag())
	rootCmd.AddCommand(webhook.NewCmdWebhook())
	rootCmd.AddCommand(newVersionCmd())
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gf version %s\n", version.Version)
		},
	}
}
