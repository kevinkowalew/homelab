package config

type Config struct {
	GithubUsername string `env:"GITHUB_USERNAME"`
	GithubToken    string `env:"GITHUB_TOKEN"`

	ArgoHost string `env:"ARGO_HOST"`
}
