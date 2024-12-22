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

	argo := argo.NewClient("https://localhost:2746")
	l := logging.New()

	p := poller.NewPoller(l, gh, argo, "#release")

	ctx := context.Background()
	if err := p.Run(ctx); err != nil {
		panic(err)
	}
}
