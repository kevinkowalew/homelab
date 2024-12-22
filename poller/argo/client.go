package argo

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host}
}

/*
func (c Client) CreateJob(name string) error {
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
*/
