<img src="https://docs.k3s.io/img/k3s-logo-dark.svg" alt="drawing" width="200"/>

# Homelab
This repo contains everything used to provision and manage my `K3s` homelab cluster.

This includes:
- an `ansible` playbook for provisioning vanilla `raspberry pi`s into `k3s` control plane nodes
- an `argo` based CI pipeline for `go` programs
- `helm` charts defining my internal services and their respective dependencies
- an `ArgoCD` app backing everything 

## CI/CD Flow
My cluster leverages a `pull` based model for CI/CD to allow the cluster to be entirely private on my local network.


- Open Repo PR --> `argo` workflow --> publish built artifact to private Docker Registry --> Post Status Check to Github --> Merge PR
- Open Homelab PR --> Bumping Service Version --> Merge --> `Argo
