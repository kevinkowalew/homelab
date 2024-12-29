package poller

import (
	"context"
	"errors"
	"fmt"
	"poller/internal/logging"
	"sort"
	"strings"
)

type Poller struct {
	logger            *logging.Logger
	scm               SCM
	ci                CI
	trigger, argoHost string
}

func NewPoller(logger *logging.Logger, scm SCM, ci CI, trigger, argoHost string) *Poller {
	return &Poller{logger, scm, ci, trigger, argoHost}
}

func (p Poller) Run(ctx context.Context) error {
	p.logger.Info("Started poller execution.")
	repos, err := p.scm.GetRepos(ctx)
	if err != nil {
		return fmt.Errorf("scm.GetRepos failed: %w", err)
	}

	shaToCIBuild, err := p.getShaToCIBuildMap(ctx)
	if err != nil {
		return err
	}

	errs := make([]error, 0)
	for _, repo := range repos {
		if err := p.handleRepo(ctx, repo, shaToCIBuild); err != nil {
			errs = append(
				errs,
				fmt.Errorf("failed to generate targets for %s/%s: %w", repo.Owner, repo.Name, err),
			)
			continue
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	p.logger.Info("Finished poller execution.")
	return nil
}

func (p Poller) handleRepo(ctx context.Context, repo Repository, shaToCIBuild map[string]CIBuild) error {
	prs, err := p.getSortedAndFilteredPRs(ctx, repo)
	if err != nil {
		return err
	}

	if len(prs) == 0 {
		return nil
	}

	shaToVersion, err := p.getShaToVersionMap(ctx, repo)
	if err != nil {
		return err
	}

	latest := Version{}
	errs := make([]error, 0)
	for _, pr := range prs {
		if pr.State == Closed {
			tagVer, ok := shaToVersion[pr.MergeSHA]
			if ok {
				latest = tagVer
				continue
			}

			build, ok := shaToCIBuild[pr.MergeSHA]
			if ok {
				latest = build.Version
				continue
			}

			latest = latest.Increment(0, 0, 1)
			// TODO: make email injectable
			if err := p.ci.CreateBuild(ctx, repo.Owner, repo.Name, latest.String(), pr.MergeSHA, "kowalewski.ke@gmail.com"); err != nil {
				errs = append(errs, err)
			}
		} else {
			commits, err := p.scm.GetPRCommits(ctx, pr)
			if err != nil {
				return fmt.Errorf("scm.GetPRCommits failed: %w", err)
			}

			for i, sha := range commits {
				tagExists := false
				tagVer, ok := shaToVersion[sha]
				if ok {
					majDiff, minDiff, patchDiff := latest.Diff(tagVer)
					if majDiff == 0 && minDiff == 0 && patchDiff == 1 {
						tagExists = true
					}
				}

				version := latest.Increment(0, 0, 1)
				version.prNumber = pr.Number
				version.buildNumber = i + 1
				tag := version.String()

				status := "success"
				description := "Build is complete"
				build, ok := shaToCIBuild[sha]
				if !ok {
					if err := p.ci.CreateBuild(ctx, repo.Owner, repo.Name, tag, sha, "kowalewski.ke@gmail.com"); err != nil {
						errs = append(errs, err)
					}

					status = "pending"
					description = "Build is in-progress"
				}

				if !tagExists {
					switch build.State {
					case Running:
						status = "pending"
						description = "Build is in-progress"
					case Failed:
						status = "failed"
						description = "Build failed"
					case Succeeded:
						status = "success"
						description = "Build finished successfully"
					}
				}

				if err := p.scm.PushStatusCheck(
					ctx,
					repo.Owner,
					repo.Name,
					sha,
					status,
					description,
					statusCheckContext,
					fmt.Sprintf("%s/workflows/argo/%s-%s-%s?tab=workflow", p.argoHost, repo.Owner, repo.Name, tag),
				); err != nil {
					return fmt.Errorf("scm.PushStatusCheck failed for %s: %w", tag, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (p Poller) getSortedAndFilteredPRs(ctx context.Context, repo Repository) ([]PR, error) {
	prs, err := p.scm.GetPRs(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("scm.GetPRs failed: %w", err)
	}

	rv := make([]PR, 0)
	for _, pr := range prs {
		if strings.Contains(pr.Description, p.trigger) {
			rv = append(rv, pr)
		}
	}

	sort.Slice(rv, func(i, j int) bool {
		if rv[i].ClosedAt != nil && rv[j].ClosedAt != nil {
			return rv[i].ClosedAt.Before(*rv[j].ClosedAt)
		}

		if rv[i].ClosedAt != nil {
			return true
		}

		return false
	})

	return rv, nil
}

func (p Poller) getShaToVersionMap(ctx context.Context, repo Repository) (map[string]Version, error) {
	tags, err := p.scm.GetTags(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("scm.GetTags failed: %w", err)
	}

	rv := make(map[string]Version, 0)
	for _, tag := range tags {
		version, err := NewVersion(tag.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag version (%s): %w", tag.Value, err)
		}
		rv[tag.SHA] = *version
	}

	return rv, nil
}

func (p Poller) getShaToCIBuildMap(ctx context.Context) (map[string]CIBuild, error) {
	builds, err := p.ci.GetCIBuilds(ctx)
	if err != nil {
		return nil, fmt.Errorf("ci.GetCIBuilds failed: %w", err)
	}

	rv := make(map[string]CIBuild)
	for _, build := range builds {
		rv[build.SHA] = build
	}

	return rv, nil
}
