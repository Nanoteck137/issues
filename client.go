package issues

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id    int    `json:"id"`
	Login string `json:"login"`
}

type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	Body      string     `json:"body"`
	HTMLURL   string     `json:"html_url"`
	Labels    []Label    `json:"labels"`
	Milestone *Milestone `json:"milestone"`
	Assignees []User     `json:"assignees"`
}

type Label struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Milestone struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
	Labels    []int64  `json:"labels,omitempty"`
}

type Client struct {
	server string
	token  string
	http   *http.Client
}

func NewClient(server, token string) *Client {
	return &Client{
		server: strings.TrimRight(server, "/"),
		token:  token,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/api/v1%s", c.server, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(bodyBytes))
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
	}

	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}
	return nil
}

type ListIssuesOptions struct {
	State     string
	Labels    []int64
	Assignee  string
	Milestone int
	Limit     int
	Page      int
}

func (c *Client) ListIssues(owner, repo string, opts ListIssuesOptions) ([]Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)

	params := url.Values{}
	if opts.State != "" {
		params.Set("state", opts.State)
	}
	if len(opts.Labels) > 0 {
		strs := make([]string, len(opts.Labels))
		for i, id := range opts.Labels {
			strs[i] = strconv.FormatInt(id, 10)
		}
		params.Set("labels", strings.Join(strs, ","))
	}
	if opts.Assignee != "" {
		params.Set("assignee", opts.Assignee)
	}
	if opts.Milestone > 0 {
		params.Set("milestone", strconv.Itoa(opts.Milestone))
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	httpReq, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	if err := c.do(httpReq, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) CreateIssue(owner, repo string, req CreateIssueRequest) (*Issue, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := c.newRequest("POST", fmt.Sprintf("/repos/%s/%s/issues", owner, repo), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var issue Issue
	if err := c.do(httpReq, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

func (c *Client) GetIssue(owner, repo string, number int) (*Issue, error) {
	httpReq, err := c.newRequest("GET", fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number), nil)
	if err != nil {
		return nil, err
	}

	var issue Issue
	if err := c.do(httpReq, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

type EditIssueRequest struct {
	State string `json:"state"`
}

func (c *Client) EditIssue(owner, repo string, number int, req EditIssueRequest) (*Issue, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := c.newRequest("PATCH", fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var issue Issue
	if err := c.do(httpReq, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) ListLabels(owner, repo string) ([]Label, error) {
	httpReq, err := c.newRequest("GET", fmt.Sprintf("/repos/%s/%s/labels", owner, repo), nil)
	if err != nil {
		return nil, err
	}

	var labels []Label
	if err := c.do(httpReq, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

func (c *Client) ListMilestones(owner, repo string) ([]Milestone, error) {
	httpReq, err := c.newRequest("GET", fmt.Sprintf("/repos/%s/%s/milestones", owner, repo), nil)
	if err != nil {
		return nil, err
	}

	var milestones []Milestone
	if err := c.do(httpReq, &milestones); err != nil {
		return nil, err
	}
	return milestones, nil
}

func (c *Client) ResolveLabels(owner, repo string, names []string) ([]int64, error) {
	if len(names) == 0 {
		return nil, nil
	}

	labels, err := c.ListLabels(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("fetching labels: %w", err)
	}

	nameToID := make(map[string]int64, len(labels))
	for _, l := range labels {
		nameToID[l.Name] = int64(l.Id)
	}

	var ids []int64
	for _, name := range names {
		id, ok := nameToID[name]
		if !ok {
			return nil, fmt.Errorf("unknown label: %q", name)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (c *Client) ResolveMilestone(owner, repo, name string) (int, error) {
	if name == "" {
		return 0, nil
	}

	milestones, err := c.ListMilestones(owner, repo)
	if err != nil {
		return 0, fmt.Errorf("fetching milestones: %w", err)
	}

	for _, m := range milestones {
		if m.Title == name {
			return m.Id, nil
		}
	}
	return 0, fmt.Errorf("milestone not found: %q", name)
}
