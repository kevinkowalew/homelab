package argo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"poller/internal/poller"
	"time"
)

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host}
}

func (c Client) CreateBuild(ctx context.Context, owner, repo, tag, sha, email string) error {
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

	workflowName := fmt.Sprintf("%s-%s-%s", owner, repo, tag)
	body := Body{
		Workflow: &Workflow{
			ApiVersion: "argoproj.io/v1alpha1",
			Kind:       "workflow",
			Metadata: Metadata{
				Name:      workflowName,
				Namespace: "argo",
				Labels: map[string]string{
					"workflows.argoproj.io/workflow-template": "github-ci-template",
				},
			},
			Spec: Spec{
				Arguments: Arguments{
					Parameters: []Parameter{
						{"repo", fmt.Sprintf("%s/%s", owner, repo)},
						{"version", tag},
						{"registry", "homelab-docker-registry:5000"},
						{"image", "golang:1.19"},
						{"revision", sha},
						{"email", email},
					},
				},
				WorkflowTemplateRef: WorkflowTemplateRef{
					Name: "github-ci-template",
				},
			},
		},
	}

	url := fmt.Sprintf("%s/api/v1/workflows/argo", c.host)
	_, err := execute[any](ctx, http.MethodPost, url, "", body)
	return err
}

func (c Client) GetCIBuilds(ctx context.Context) ([]poller.CIBuild, error) {
	type (
		Item struct {
			Spec struct {
				Arguments struct {
					Parameters []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"parameters"`
				} `json:"arguments"`
			} `json:"spec"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		}
		response struct {
			Items []Item `json:"items"`
		}
	)

	url := fmt.Sprintf("%s/api/v1/workflows/argo", c.host)
	res, err := execute[response](ctx, http.MethodGet, url, "", nil)
	if err != nil {
		return nil, err
	}

	rv := make([]poller.CIBuild, 0)
	for _, wf := range res.Items {
		get := func(i Item, key string) string {
			for _, param := range i.Spec.Arguments.Parameters {
				if param.Name == key {
					return param.Value
				}
			}

			return ""
		}

		normalized := poller.CIBuild{
			SHA: get(wf, "revision"),
		}

		switch wf.Status.Phase {
		case "Running":
			normalized.State = poller.Running
		case "Failed":
			normalized.State = poller.Failed
		case "Succeeded":
			normalized.State = poller.Succeeded
		}

		version, err := poller.NewVersion(get(wf, "version"))
		if err != nil {
			return nil, err
		}
		normalized.Version = *version
		rv = append(rv, normalized)
	}

	return rv, nil
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
		return nil, err
	}
	return t, nil
}
