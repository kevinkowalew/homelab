# Setting up Argo
```sh
kubectl create namespace argocd
```

```sh
kubectl apply -n argocd -f install.yaml
```

```sh
kubectl port-forward -n argocd svc/argocd-server 8080:443
```

```sh
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d | pbcopy
```

```sh
kubectl apply -f argo-cd/app.yaml -n argocd
```

# Troubleshooting
If `crd/applications.argoproj.io` gets stuck during deletion run this command:
```sh
kubectl patch crd/applications.argoproj.io -p '{"metadata":{"finalizers":[]}}' --type=merge
```
