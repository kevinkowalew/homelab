# Setting up Argo
```sh
kubectl create namespace argocd
```

```sh
kubectl apply -n argocd -f install.yaml
```

```sh
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d | pbcopy
```

```sh
kubectl patch crd/applications.argoproj.io -p '{"metadata":{"finalizers":[]}}' --type=merge
```

```sh
kubectl apply -f argo-cd/app.yaml -n argocd
```
