package poller

import (
	"context"
	"time"
)

const (
	Open   PRState = "open"
	Closed PRState = "closed"

	Running   CIBuildState = "Running"
	Failed    CIBuildState = "Failed"
	Succeeded CIBuildState = "Succeeded"

	Success PRStatusCheckState = "success"
	Failure PRStatusCheckState = "failure"
	Pending PRStatusCheckState = "pending"

	statusCheckContext = "continuous-integration/argo"
)

type (
	Repository struct {
		Owner, Name string
	}

	PRState string
	PR      struct {
		TargetBranch string
		Number       int
		Description  string
		State        PRState
		MergeSHA     string
		ClosedAt     *time.Time
	}

	PRStatusCheckState string
	PRStatusCheck      struct {
		State                                PRStatusCheckState
		URL, Description, Context, TargetURL string
	}

	Tag struct {
		Value, SHA string
	}

	Build struct {
		SHA string
	}

	SCM interface {
		GetRepos(context.Context) ([]Repository, error)
		GetPRs(context.Context, Repository) ([]PR, error)
		GetTags(context.Context, Repository) ([]Tag, error)
		GetPRCommits(context.Context, PR) ([]string, error)
		PushStatusCheck(ctx context.Context, owner, repo, sha, state, description, context, targetURL string) error
		GetStatusChecks(ctx context.Context, repo Repository, sha string) ([]PRStatusCheck, error)
	}

	CIBuildState string
	CI           interface {
		GetCIBuilds(context.Context) ([]CIBuild, error)
		CreateBuild(ctx context.Context, owner, name, tag, sha, email string) error
	}

	CIBuild struct {
		SHA     string
		State   CIBuildState
		Version Version
	}
)
