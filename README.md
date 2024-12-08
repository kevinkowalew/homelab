<img src="https://docs.k3s.io/img/k3s-logo-dark.svg" alt="drawing" width="200"/>
# Homelab
This repo contains everything used to provision and manage my `K3s` homelab cluster.
I leverage GitOps

## Provisioning
- I use `ansible` to provision my control plane nodes

## CI/CD
- I use `argo` events and workflows to manage CI for my cluster.
### Flow
- Open Repo PR --> `argo` workflow --> publish built artifact to private Docker Registry --> Post Status Check to Github --> Merge PR
- Open Homelab PR --> Bumping Service Version --> Merge --> `Argo

