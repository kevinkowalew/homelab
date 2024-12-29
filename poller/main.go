package main

import (
	"context"
	"os"
	"poller/argo"
	"poller/github"
	"poller/internal/logging"
	"poller/internal/poller"
)

func main() {
	gh := github.New(
		os.Getenv("GITHUB_TOKEN"),
		os.Getenv("GITHUB_USERNAME"),
	)

	argoHost := "https://localhost:2746"
	argo := argo.NewClient(argoHost)
	l := logging.New()

	p := poller.NewPoller(l, gh, argo, "#release", argoHost)

	ctx := context.Background()
	if err := p.Run(ctx); err != nil {
		panic(err)
	}
}
