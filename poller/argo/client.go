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
	"strings"
	"time"
)

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host}
}

func (c Client) CreateBuild(ctx context.Context, repo poller.Repository, build poller.Build) error {
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

	name := fmt.Sprintf("%s-%s-%s", repo.Owner, repo.Name, build.Tag())
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
						{"repo", fmt.Sprintf("%s/%s", repo.Owner, repo.Name)},
						{"version", build.Tag()},
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
	_, err := execute[any](ctx, http.MethodPost, url, "", body)
	return err
}

func (c Client) GetCIBuildStatus(ctx context.Context, repo poller.Repository, build poller.Build) (poller.CIBuildState, error) {
	type response struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"phase"`
	}
	name := fmt.Sprintf("%s-%s-%s", repo.Owner, repo.Name, build.Tag())
	url := "https://localhost:2746/api/v1/workflows/argo/" + name
	res, err := execute[response](ctx, http.MethodGet, url, "", nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), "404 Not Found:") {
			return poller.NotStarted, nil
		} else {
			return "", err
		}
	}

	switch res.Status.Phase {
	case "Running":
		return poller.Running, nil
	case "Failed":
		return poller.Failed, nil
	case "Succeeded":
		return poller.Succeeded, nil
	default:
		return "", nil
	}
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
		fmt.Println("weird: " + err.Error())
		return nil, err
	}
	return t, nil
}
