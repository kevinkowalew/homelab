package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type (
	Repository struct {
		Name string `json:"full_name"`
	}

	PullRequest struct {
		Number         int      `json:"number"`
		Body           string   `json:"body"`
		State          string   `json:"state"`
		CommitsUrl     string   `json:"commits_url"`
		Head           *Head    `json:"head"`
		MergeCommitSHA string   `json:"merge_commit_sha"`
		Commits        *Commits `json:"commits"`
	}

	Commits struct {
		Href string `json:"href"`
	}

	Head struct {
		Branch string `json:"ref"`
		SHA    string `json:"sha"`
		Repo   *Repo  `json:"repo"`
	}

	Repo struct {
		DefaultBranch string `json:"default_branch"`
	}

	Tag struct {
		Name   string  `json:"name"`
		Commit *Commit `json:"commit"`
	}

	TagRef struct {
		Ref    string `json:"ref"`
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	Commit struct {
		SHA    string `json:"sha"`
		Commit struct {
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	Github struct {
		token, username string
	}

	BuildRequest struct {
		Repository
		SHA     string
		Version Version
	}

	Version struct {
		Major, Minor, Patch, BuildNumber int
		PreRelease                       bool
	}
)

func main() {
	name := "my-job-1"
	if err := getArgoJobStatus(name); err != nil {
		panic(err)
	}
}

func main1() {
	client := New(
		os.Getenv("GITHUB_TOKEN"),
		os.Getenv("GITHUB_USERNAME"),
	)

	ctx := context.Background()
	repos, err := client.GetRepos(ctx)
	if err != nil {
		panic(err)
	}

	for _, repo := range *repos {
		_, err := client.GetPullRequests(ctx, repo.Name)
		if err != nil {
			panic(err)
		}

		lastRelease, err := client.GetLatestVerion(ctx, repo.Name)
		if err != nil {
			panic(err)
		}

		tags, err := client.GetTags(ctx, repo.Name)
		if err != nil {
			panic(err)
		}

		commitToTag := make(map[string]*Version, 0)
		for _, tag := range *tags {
			if tag.Commit != nil && tag.Commit.SHA != "" {
				version, err := NewVersion(tag.Name)
				if err != nil {
					fmt.Println("failed to parse version tag: " + err.Error())
					continue
				}
				commitToTag[tag.Commit.SHA] = version
			}
		}

		prs, err := client.GetPullRequests(ctx, repo.Name)
		if len(*prs) > 0 {
			for _, pr := range *prs {
				if !strings.Contains(pr.Body, "#release") {
					continue
				}

				if pr.State == "closed" {
					_, ok := commitToTag[pr.MergeCommitSHA]
					if !ok {
						fmt.Println(BuildRequest{repo, pr.Head.SHA, lastRelease.Clone(0, 0, 1, 0, false)})
					}
				} else if pr.State == "open" {
					commits, err := execute[[]Commit](ctx, http.MethodGet, pr.CommitsUrl, os.Getenv("GITHUB_TOKEN"), nil)
					if err != nil {
						fmt.Println("failed to fetch commits for PR: " + err.Error())
						continue
					}

					sort.Slice(*commits, func(i, j int) bool {
						return (*commits)[i].Commit.Committer.Date.Before((*commits)[j].Commit.Committer.Date)
					})

					for i, commit := range *commits {
						_, ok := commitToTag[commit.SHA]
						if !ok {
							fmt.Println(BuildRequest{repo, commit.SHA, lastRelease.Clone(0, 0, 1, i+1, true)})
						}
					}

				}
			}
		}

	}
}

func New(token, username string) *Github {
	return &Github{token, username}
}

func (g Github) GetRepos(ctx context.Context) (*[]Repository, error) {
	url := "https://api.github.com/user/repos?affiliation=owner"
	return execute[[]Repository](ctx, http.MethodGet, url, g.token, nil)
}

func (g Github) GetPullRequests(ctx context.Context, name string) (*[]PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/pulls?state=all", name)
	return execute[[]PullRequest](ctx, http.MethodGet, url, g.token, nil)
}

func (g Github) PushStatusCheck(ctx context.Context, name, sha, state, description, context string) error {
	body := struct {
		State       string `json:"state"`
		Description string `json:"description"`
		Context     string `json:"context"`
	}{state, description, context}

	url := fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s", name, sha)
	_, err := execute[any](ctx, http.MethodPost, url, g.token, body)
	return err
}

func (g Github) CreateTag(ctx context.Context, name, commit, tag string) error {
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
	res, err := execute[response](ctx, http.MethodPost, url, g.token, body)
	if err != nil {
		return err
	}

	refBody := struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	}{"refs/tags/" + tag, res.SHA}
	url = fmt.Sprintf("https://api.github.com/repos/%s/git/refs", name)
	_, err = execute[any](ctx, http.MethodPost, url, g.token, refBody)
	return err
}

func (g Github) DeleteTag(ctx context.Context, name, tag string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags/%s", name, tag)
	_, err := execute[struct{}](ctx, http.MethodDelete, url, g.token, nil)
	return err
}

func (g Github) GetTags(ctx context.Context, name string) (*[]Tag, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", name)
	return execute[[]Tag](ctx, http.MethodGet, url, g.token, nil)
}

func (g Github) GetLatestVerion(ctx context.Context, name string) (*Version, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", name)
	tags, err := execute[[]Tag](ctx, http.MethodGet, url, g.token, nil)
	if err != nil {
		return nil, err
	}

	versions := make([]Version, 0)
	for _, tag := range *tags {
		version, err := NewVersion(tag.Name)
		if err != nil {
			// TODO: make logging handling good here
			fmt.Println(err.Error())
		}

		if !version.PreRelease {
			versions = append(versions, *version)
		}
	}

	if len(versions) == 0 {
		return &Version{0, 0, 0, 0, false}, nil
	}

	sort.Slice(versions, func(i, j int) bool {
		if versions[i].Major < versions[j].Major {
			return true
		} else if versions[i].Minor < versions[j].Minor {
			return true
		} else if versions[i].Patch < versions[j].Patch {
			return true
		}

		return false
	})

	return &versions[len(versions)-1], nil
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

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Do(req)
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
		fmt.Println("weird: " + err.Error())
		return nil, err
	}
	return t, nil
}

func NewVersion(s string) (*Version, error) {
	if !strings.HasPrefix(s, "v") {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	s = s[1:]

	p := strings.Split(s, ".")
	if len(p) != 3 {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	major, err := strconv.Atoi(p[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse major version tag: %w", err)
	}

	minor, err := strconv.Atoi(p[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor version tag: %w", err)
	}

	pp := strings.Split(p[2], "-")
	patch, err := strconv.Atoi(pp[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch version tag: %w", err)
	}

	if len(pp) != 1 && len(pp) != 3 {
		return nil, fmt.Errorf("unexpected version tag formatting: %s", s)
	}

	if len(pp) == 1 {
		return &Version{major, minor, patch, 0, false}, nil
	}

	// TODO: make the prerelease determination here more stable
	return &Version{major, minor, patch, 0, len(pp) == 4}, nil
}

func (v Version) Clone(major, minor, patch int, buildNumber int, preRelease bool) Version {
	return Version{
		Major:       v.Major + major,
		Minor:       v.Minor + minor,
		Patch:       v.Patch + patch,
		PreRelease:  preRelease,
		BuildNumber: buildNumber,
	}
}

func (v Version) String() string {
	if v.PreRelease {
		return fmt.Sprintf("v%d.%d.%d-prerelease-%d", v.Major, v.Minor, v.Patch, v.BuildNumber)
	}

	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func createArgoJob(name string) error {
	type (
		Metadata struct {
			Name      string            `json:"name"`
			Namespace string            `json:"namespace"`
			Labels    map[string]string `json:"labels"`
		}

		Parameter struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}

		Arguments struct {
			Parameters []Parameter `json:"parameters"`
		}

		WorkflowTemplateRef struct {
			Name string `json:"name"`
		}

		Spec struct {
			Arguments           Arguments           `json:"arguments"`
			WorkflowTemplateRef WorkflowTemplateRef `json:"workflowTemplateRef"`
		}

		Workflow struct {
			ApiVersion string   `json:"apiVersion"`
			Kind       string   `json:"kind"`
			Metadata   Metadata `json:"metadata"`
			Spec       Spec     `json:"spec"`
		}

		Body struct {
			Workflow *Workflow `json:"workflow"`
		}
	)

	body := Body{
		Workflow: &Workflow{
			ApiVersion: "argoproj.io/v1alpha1",
			Kind:       "workflow",
			Metadata: Metadata{
				Name:      name,
				Namespace: "argo",
				Labels: map[string]string{
					"workflows.argoproj.io/workflow-template": "github-ci-template",
				},
			},
			Spec: Spec{
				Arguments: Arguments{
					Parameters: []Parameter{
						{"repo", "kevinkowalew/auth-server"},
						{"version", "0.0.3"},
						{"registry", "homelab-docker-registry:5000"},
						{"image", "golang:1.19"},
					},
				},
				WorkflowTemplateRef: WorkflowTemplateRef{
					Name: "github-ci-template",
				},
			},
		},
	}

	url := "https://localhost:2746/api/v1/workflows/argo"
	ctx := context.Background()
	_, err := execute[any](ctx, http.MethodPost, url, "", body)
	return err
}

func getArgoJobStatus(jobName string) error {
	type response struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"phase"`
	}
	url := "https://localhost:2746/api/v1/workflows/argo/" + jobName
	ctx := context.Background()
	_, err := execute[response](ctx, http.MethodPost, url, "", nil)
	if err != nil {
		panic(err)
	}

	return err

	// possible job states
	// "Running"
	// "Failed"
	// "Succeeded"
}
