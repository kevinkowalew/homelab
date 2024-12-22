package poller

import (
	"context"
	"errors"
	"fmt"
	"poller/internal/logging"
	"strings"
	"time"
)

const (
	Open   PRState = "open"
	Closed PRState = "closed"

	NotStarted CIBuildState = "NotStarted"
	Running    CIBuildState = "Running"
	Failed     CIBuildState = "Failed"
	Succeeded  CIBuildState = "Succeeded"
)

type (
	Repository struct {
		Owner, Name string
	}

	PRState string
	PR      struct {
		Parent       Repository
		TargetBranch string
		Number       int
		Description  string
		State        PRState
		MergeSHA     string
	}

	Tag struct {
		Value, SHA string
		Timestamp  time.Time
	}

	Build struct {
		pr          PR
		sha         string
		version     version
		prCommitNum int
	}

	SCM interface {
		GetRepos(context.Context) ([]Repository, error)
		GetPRs(context.Context, Repository) ([]PR, error)
		GetTags(context.Context, Repository) ([]Tag, error)
		GetPRCommits(context.Context, PR) ([]string, error)
	}

	CIBuildState string
	CI           interface {
		GetCIBuildStatus(context.Context, Repository, Build) (CIBuildState, error)
		CreateBuild(context.Context, Repository, Build) error
	}
)

type Poller struct {
	logger  *logging.Logger
	scm     SCM
	ci      CI
	trigger string
}

func NewPoller(logger *logging.Logger, scm SCM, ci CI, trigger string) *Poller {
	return &Poller{logger, scm, ci, trigger}
}

func (p Poller) Run(ctx context.Context) error {
	repos, err := p.scm.GetRepos(ctx)
	if err != nil {
		return fmt.Errorf("scm.GetRepos failed: %w", err)
	}

	errs := make([]error, 0)
	for _, repo := range repos {
		targets, err := p.generateTargets(ctx, repo)
		if err != nil {
			p.logger.Error(err, "failed to generate build targets", "repo_owner", repo.Owner, "repo_name", repo.Name)
			continue
		}

		for _, target := range targets {
			status, err := p.ci.GetCIBuildStatus(ctx, repo, target)
			if err != nil {
				p.logger.Error(err, "ci.GetCIBuildStatus failed", "tag", target.Tag())
				continue
			}

			if status == NotStarted {
				if err := p.ci.CreateBuild(ctx, repo, target); err != nil {
					p.logger.Error(err, "ci.CreateBuild failed", "tag", target.Tag())
				}
			} else if status == Running {
				p.logger.Info("Skipping CI build as it's already running.", "tag", target.Tag())
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (p Poller) generateTargets(ctx context.Context, repo Repository) ([]Build, error) {
	l := p.logger.WithFields("repo_owner", repo.Owner, "repo_name", repo.Name)
	prs, err := p.scm.GetPRs(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("scm.GetPRs failed: %w", err)
	}

	tags, err := p.scm.GetTags(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("scm.GetTags failed: %w", err)
	}

	latest := version{0, 0, 0}
	tagShaToVersion := make(map[string]version, 0)
	for _, tag := range tags {
		version, err := newVersion(tag.Value)
		if err != nil {
			l.Error(err, "failed to parse version", "tag_value")
			continue
		}
		tagShaToVersion[tag.SHA] = *version

		if latest.Less(*version) {
			latest = *version
		}
	}

	if len(prs) == 0 {
		return nil, nil
	}

	bts := make([]Build, 0)
	for _, pr := range prs {
		if !strings.Contains(pr.Description, p.trigger) {
			continue
		}

		if pr.State == Closed && pr.MergeSHA != "" {
			_, ok := tagShaToVersion[pr.MergeSHA]
			if !ok {
				bts = append(bts, Build{
					sha:     pr.MergeSHA,
					version: latest.Increment(0, 0, 1),
					pr:      pr,
				})
			}
			continue
		}

		if pr.State == Open {
			commits, err := p.scm.GetPRCommits(ctx, pr)
			if err != nil {
				l.Error(err, "scm.GetPRCommits failed", "pr_number", pr.Number)
				continue
			}

			for i, sha := range commits {
				_, ok := tagShaToVersion[sha]
				if !ok {
					bts = append(bts, Build{
						sha:         sha,
						version:     latest.Increment(0, 0, 1),
						pr:          pr,
						prCommitNum: i + 1,
					})
				}
			}

		}
	}

	return bts, nil
}

func (t Build) Tag() string {
	if t.sha == t.pr.MergeSHA {
		return fmt.Sprintf("v%d.%d.%d", t.version.major, t.version.minor, t.version.patch)
	}

	return fmt.Sprintf("v%d.%d.%d-pr-%d-build-%d", t.version.major, t.version.minor, t.version.patch, t.pr.Number, t.prCommitNum)
}
