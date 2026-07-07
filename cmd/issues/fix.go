package main

import (
	"fmt"
	"strings"

	"github.com/nanoteck137/issues"
	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Fix issues with literal \\n in their body",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := resolveConfig(cmd)

		if cfg.Repo == "" {
			return fmt.Errorf("--repo is required")
		}

		client := issues.NewClient(cfg.Server, cfg.Token)

		opts := issues.ListIssuesOptions{
			State: "all",
		}

		issuesList, err := client.ListIssues(cfg.Owner, cfg.Repo, opts)
		if err != nil {
			return fmt.Errorf("listing issues: %w", err)
		}

		var fixed int
		for _, issue := range issuesList {
			if !strings.Contains(issue.Body, "\\n") {
				continue
			}

			newBody := strings.ReplaceAll(issue.Body, "\\n", "\n")

			req := issues.EditIssueRequest{Body: newBody}
			_, err := client.EditIssue(cfg.Owner, cfg.Repo, issue.Number, req)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Failed to fix #%d: %v\n", issue.Number, err)
				continue
			}

			fmt.Printf("Fixed issue #%d\n", issue.Number)
			fixed++
		}

		fmt.Printf("Done. Fixed %d issues.\n", fixed)
		return nil
	},
}

func init() {
	fixCmd.Flags().StringVarP(&repoFlag, "repo", "r", gitInfo.Repo, "Repository name")

	rootCmd.AddCommand(fixCmd)
}
