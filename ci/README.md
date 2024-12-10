# CI
- `argo workflows` power CI in my cluster.

## Inputs
| Name     | Description                                              | Example Value        |
| -------- | -------------------------------------------------------- | -------------------- |
| repo     | github repository you wish to target                     | kevinkowalew/go-api  |
| version  | artifact verison you wish to push                        | 0.0.1                |
| registry | docker registry address                                  | docker-registry:5000 |
| image    | base image to use for build, lint & test step containers | golang:1.19          |
| test     | shell command to use in build step                       | go build -o api .    |
| build    | shell command to use in test step                        | go build ./...       |
