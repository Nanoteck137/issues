package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nanoteck137/issues"
	"github.com/spf13/cobra"
)

var (
	repoFlag      string
	titleFlag     string
	bodyFlag      string
	bodyFileFlag  string
	editorFlag    bool
	labelFlags    []string
	milestoneFlag string
	assigneeFlags []string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := resolveConfig(cmd)

		if cfg.Repo == "" {
			return fmt.Errorf("--repo is required")
		}

		title := titleFlag
		body, err := resolveBody(title)
		if err != nil {
			return err
		}

		if title == "" {
			return fmt.Errorf("--title is required (or provide a title via --editor)")
		}

		client := issues.NewClient(cfg.Server, cfg.Token)

		labelIDs, err := client.ResolveLabels(cfg.Owner, cfg.Repo, labelFlags)
		if err != nil {
			return fmt.Errorf("resolving labels: %w", err)
		}

		milestoneID, err := client.ResolveMilestone(cfg.Owner, cfg.Repo, milestoneFlag)
		if err != nil {
			return fmt.Errorf("resolving milestone: %w", err)
		}

		if body != "" {
			body += "\n\n---\n"
		}
		body += "_Created by issues_"

		req := issues.CreateIssueRequest{
			Title:     title,
			Body:      body,
			Assignees: assigneeFlags,
			Milestone: milestoneID,
			Labels:    labelIDs,
		}

		issue, err := client.CreateIssue(cfg.Owner, cfg.Repo, req)
		if err != nil {
			return fmt.Errorf("creating issue: %w", err)
		}

		if cfg.JSON {
			out := struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
				URL    string `json:"url"`
			}{
				Number: issue.Number,
				Title:  issue.Title,
				URL:    issue.HTMLURL,
			}
			return json.NewEncoder(os.Stdout).Encode(out)
		}

		fmt.Printf("✓ Created issue #%d\n\n%s\n\n%s\n", issue.Number, issue.Title, issue.HTMLURL)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&repoFlag, "repo", "r", gitInfo.Repo, "Repository name")
	createCmd.Flags().StringVarP(&titleFlag, "title", "t", "", "Issue title")
	createCmd.Flags().StringVarP(&bodyFlag, "body", "b", "", "Issue body")
	createCmd.Flags().StringVar(&bodyFileFlag, "body-file", "", "Read issue body from file")
	createCmd.Flags().BoolVar(&editorFlag, "editor", false, "Open editor to write issue")
	createCmd.Flags().StringSliceVarP(&labelFlags, "label", "l", nil, "Labels (by name)")
	createCmd.Flags().StringVarP(&milestoneFlag, "milestone", "m", "", "Milestone (by name)")
	createCmd.Flags().StringSliceVarP(&assigneeFlags, "assignee", "a", nil, "Assignees (by username)")

	rootCmd.AddCommand(createCmd)
}

func isStdinPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

func resolveBody(title string) (string, error) {
	switch {
	case bodyFlag != "":
		return bodyFlag, nil

	case bodyFileFlag != "":
		data, err := os.ReadFile(bodyFileFlag)
		if err != nil {
			return "", fmt.Errorf("reading body file: %w", err)
		}
		return string(data), nil

	case isStdinPiped():
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil

	case editorFlag:
		t, body, err := editIssue(title)
		if err != nil {
			return "", err
		}
		titleFlag = t
		return body, nil

	default:
		return "", nil
	}
}

func editIssue(existingTitle string) (title, body string, err error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "issue-*.md")
	if err != nil {
		return "", "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	var initial string
	if existingTitle != "" {
		initial = "# " + existingTitle + "\n\n"
	} else {
		initial = "# Title\n\n"
	}

	if _, err := tmpFile.WriteString(initial); err != nil {
		tmpFile.Close()
		return "", "", fmt.Errorf("writing template: %w", err)
	}
	tmpFile.Close()

	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return "", "", fmt.Errorf("editor exited with error: %w", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", "", fmt.Errorf("reading edited file: %w", err)
	}

	parsedTitle, parsedBody := parseEditorContent(string(data))
	return parsedTitle, parsedBody, nil
}

func parseEditorContent(content string) (title, body string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ""
	}

	lines := strings.SplitN(content, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])

	if strings.HasPrefix(firstLine, "# ") {
		title = strings.TrimSpace(strings.TrimPrefix(firstLine, "# "))
		if len(lines) > 1 {
			body = strings.TrimSpace(lines[1])
		}
	} else {
		body = content
	}
	return
}
