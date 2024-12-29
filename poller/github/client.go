package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"poller/internal/poller"
	"sort"
	"time"
)

type Client struct {
	token, username string
}

func New(token, username string) *Client {
	return &Client{token, username}
}

func (c Client) GetRepos(ctx context.Context) ([]poller.Repository, error) {
	type repository struct {
		Name string `json:"name"`
	}
	url := "https://api.github.com/user/repos?affiliation=owner"
	repos, err := execute[[]repository](ctx, http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}

	rv := make([]poller.Repository, 0)
	for _, repo := range *repos {
		rv = append(rv, poller.Repository{
			Owner: c.username,
			Name:  repo.Name,
		})
	}

	return rv, nil
}

func (c Client) GetPRs(ctx context.Context, repo poller.Repository) ([]poller.PR, error) {
	type pr struct {
		Number   int        `json:"number"`
		Body     string     `json:"body"`
		State    string     `json:"state"`
		ClosedAt *time.Time `json:"closed_at"`
		Head     struct {
			SHA  string `json:"sha"`
			Repo struct {
				DefaultBranch string `json:"default_branch"`
			} `json:"repo"`
		} `json:"head"`
		MergeCommitSHA string `json:"merge_commit_sha"`
		Commits        struct {
			Href string `json:"href"`
		} `json:"commits"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all", repo.Owner, repo.Name)
	prs, err := execute[[]pr](ctx, http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}

	rv := make([]poller.PR, 0)
	for _, pr := range *prs {
		normalized := poller.PR{
			Number:       pr.Number,
			Description:  pr.Body,
			MergeSHA:     pr.MergeCommitSHA,
			TargetBranch: pr.Head.Repo.DefaultBranch,
			ClosedAt:     pr.ClosedAt,
		}

		if pr.State == "open" {
			normalized.State = poller.Open
		} else if pr.State == "closed" {
			normalized.State = poller.Closed
		}
		rv = append(rv, normalized)

	}

	return rv, nil
}
func (c Client) GetTags(ctx context.Context, repo poller.Repository) ([]poller.Tag, error) {
	type tag struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", repo.Owner, repo.Name)
	tags, err := execute[[]tag](ctx, http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}

	rv := make([]poller.Tag, 0)
	for _, tag := range *tags {
		rv = append(rv, poller.Tag{
			Value: tag.Name,
			SHA:   tag.Commit.SHA,
		})
	}

	return rv, nil
}

func (c Client) GetPRCommits(ctx context.Context, pr poller.PR) ([]string, error) {
	type commit struct {
		Name   string `json:"name"`
		SHA    string `json:"sha"`
		Commit struct {
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	url := fmt.Sprintf("https://api.github.com/repos/kevinkowalew/apigateway/pulls/%d/commits", pr.Number)
	commits, err := execute[[]commit](ctx, http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}

	sort.Slice(*commits, func(i, j int) bool {
		return (*commits)[i].Commit.Committer.Date.Before((*commits)[j].Commit.Committer.Date)
	})

	rv := make([]string, 0)
	for _, commit := range *commits {
		rv = append(rv, commit.SHA)
	}

	return rv, nil
}

func (c Client) PushStatusCheck(ctx context.Context, owner, repo, sha, state, description, context, targetURL string) error {
	body := struct {
		State       string `json:"state"`
		Description string `json:"description"`
		Context     string `json:"context"`
		TargetURL   string `json:"target_url"`
	}{state, description, context, targetURL}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s", owner, repo, sha)
	_, err := execute[any](ctx, http.MethodPost, url, c.token, body)
	return err
}

func (c Client) GetStatusChecks(ctx context.Context, repo poller.Repository, sha string) ([]poller.PRStatusCheck, error) {
	type (
		status struct {
			URL         string `json:"url"`
			State       string `json:"state"`
			Description string `json:"description"`
			Context     string `json:"context"`
			TargetURL   string `json:"target_url"`
		}
		response struct {
			State    string   `json:"state"`
			Statuses []status `json:"statuses"`
		}
	)

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s/status", repo.Owner, repo.Name, sha)
	res, err := execute[response](ctx, http.MethodGet, url, c.token, nil)
	if err != nil {
		return nil, err
	}

	rv := make([]poller.PRStatusCheck, 0)
	for _, status := range res.Statuses {
		normalized := poller.PRStatusCheck{
			URL:         status.URL,
			Description: status.Description,
			Context:     status.Context,
			TargetURL:   status.TargetURL,
		}

		switch status.State {
		case "error":
		case "failure":
			normalized.State = poller.Failure
		case "pending":
			normalized.State = poller.Pending
		case "success":
			normalized.State = poller.Success
		}

		rv = append(rv, normalized)
	}
	return rv, err
}

func (c Client) CreateTag(ctx context.Context, name, commit, tag string) error {
	body := struct {
		Tag     string `json:"tag"`
		Message string `json:"message"`
		Object  string `json:"object"`
		Type    string `json:"type"`
	}{name, "ArgoCI pushing pre-release tag", commit, "commit"}

	type response struct {
		SHA string `json:"sha"`
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/tags", name)
	res, err := execute[response](ctx, http.MethodPost, url, c.token, body)
	if err != nil {
		return err
	}

	refBody := struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	}{"refs/tags/" + tag, res.SHA}
	url = fmt.Sprintf("https://api.github.com/repos/%s/git/refs", name)
	_, err = execute[any](ctx, http.MethodPost, url, c.token, refBody)
	return err
}

func (c Client) DeleteTag(ctx context.Context, name, tag string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags/%s", name, tag)
	_, err := execute[struct{}](ctx, http.MethodDelete, url, c.token, nil)
	return err
}

func execute[T any](ctx context.Context, verb, url, token string, body any) (*T, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal failed: %w", err)
		}
		r = bytes.NewBuffer(b)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, verb, url, r)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		msg := fmt.Sprintf("%s: %s", res.Status, body)
		return nil, errors.New(msg)
	}

	var t *T
	if string(resBody) == "" {
		return nil, nil
	}
	err = json.Unmarshal(resBody, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
