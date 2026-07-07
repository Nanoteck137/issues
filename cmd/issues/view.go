package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nanoteck137/issues"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view <issue-number>",
	Short: "Show details of a specific issue",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := resolveConfig(cmd)

		number, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number: %s", args[0])
		}

		if cfg.Repo == "" {
			return fmt.Errorf("--repo is required")
		}

		client := issues.NewClient(cfg.Server, cfg.Token)

		issue, err := client.GetIssue(cfg.Owner, cfg.Repo, number)
		if err != nil {
			return fmt.Errorf("fetching issue: %w", err)
		}

		if cfg.JSON {
			return json.NewEncoder(os.Stdout).Encode(issue)
		}

		fmt.Printf("#%d %s\n", issue.Number, issue.Title)
		fmt.Printf("State: %s\n", issue.State)

		if len(issue.Labels) > 0 {
			names := make([]string, len(issue.Labels))
			for i, l := range issue.Labels {
				names[i] = l.Name
			}
			fmt.Printf("Labels: %s\n", strings.Join(names, ", "))
		}

		if issue.Milestone != nil {
			fmt.Printf("Milestone: %s\n", issue.Milestone.Title)
		}

		if len(issue.Assignees) > 0 {
			names := make([]string, len(issue.Assignees))
			for i, a := range issue.Assignees {
				names[i] = a.Login
			}
			fmt.Printf("Assignees: %s\n", strings.Join(names, ", "))
		}

		if issue.Body != "" {
			fmt.Printf("\n%s\n", issue.Body)
		}

		fmt.Printf("\n%s\n", issue.HTMLURL)
		return nil
	},
}

func init() {
	viewCmd.Flags().StringVarP(&repoFlag, "repo", "r", gitInfo.Repo, "Repository name")

	rootCmd.AddCommand(viewCmd)
}
