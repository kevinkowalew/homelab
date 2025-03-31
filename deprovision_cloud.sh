helm delete argo --namespace argo
kubectl delete -f helm/app.yaml
kubectl delete -f helm/ci-workflow.yaml
kubectl delete secret git-creds -n argo
