package cmd

import (
	"fmt"
	"os"

	"github.com/josinSbazin/gf/cmd/auth"
	"github.com/josinSbazin/gf/cmd/mr"
	"github.com/josinSbazin/gf/cmd/pipeline"
	"github.com/josinSbazin/gf/cmd/repo"
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(auth.NewCmdAuth())
	rootCmd.AddCommand(mr.NewCmdMR())
	rootCmd.AddCommand(pipeline.NewCmdPipeline())
	rootCmd.AddCommand(repo.NewCmdRepo())
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
