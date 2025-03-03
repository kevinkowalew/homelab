# CI
- `argo workflows` power CI for my homelab.  Workflows can be triggered with the following


## Inputs
| Name     | Description                                              | Example Value        |
| -------- | -------------------------------------------------------- | -------------------- |
| repo     | github repository you wish to target                     | kevinkowalew/go-api  |
| version  | docker image verison you wish to push                    | v0.0.1               |
| registry | docker registry address                                  | docker-registry:5000 |
| revision | commit hash SHA you wish to build                        |                      |
| image    | base image to use for build, lint & test step containers | golang:1.23          |

# CD

## Get Admin Password
```sh
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d | pbcopy
```
