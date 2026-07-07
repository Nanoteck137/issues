package issues

import (
	"net/url"
	"os/exec"
	"strings"
)

type GitInfo struct {
	Server string
	Owner  string
	Repo   string
}

func DetectFromGit() GitInfo {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return GitInfo{}
	}
	return ParseGitURL(strings.TrimSpace(string(out)))
}

func ParseGitURL(raw string) GitInfo {
	raw = strings.TrimSuffix(raw, ".git")

	var host, path string

	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return GitInfo{}
		}
		host = u.Host
		path = strings.TrimPrefix(u.Path, "/")
	} else if strings.Contains(raw, "@") {
		parts := strings.SplitN(raw, "@", 2)
		if len(parts) != 2 {
			return GitInfo{}
		}
		rest := parts[1]
		colonIdx := strings.Index(rest, ":")
		if colonIdx < 0 {
			return GitInfo{}
		}
		host = rest[:colonIdx]
		path = rest[colonIdx+1:]
	} else {
		return GitInfo{}
	}

	slashIdx := strings.Index(path, "/")
	if slashIdx < 0 {
		return GitInfo{}
	}

	return GitInfo{
		Server: "https://" + host,
		Owner:  path[:slashIdx],
		Repo:   path[slashIdx+1:],
	}
}
