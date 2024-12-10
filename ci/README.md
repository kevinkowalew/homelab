# CI
- `argo workflows` power CI in my cluster.

## Inputs
| Name                                        | Description                                                                                                                                                                                 | Example                                                                                  |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| repo

version

registry

image

build

test | <br>

github repository you wish to target

image version

docker registry

base image to use for build, lint and test steps

command used for build step

command used for test step

<br> | kevinkowalew/go-api

0.0.1

registry:5000

golang:1.23

go build -o api .

go test ./... |
