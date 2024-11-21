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
