kubectl delete -f helm/app.yaml
kubectl delete -f helm/ci-workflow.yaml
kubectl delete cm argo-pr-poller -n argo
kubectl delete cj argo-pr-poller -n argo
kubectl delete secret git-creds -n argo
kubectl delete namespace argo
kubectl delete namespace prod
