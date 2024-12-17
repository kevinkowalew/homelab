kubectl delete namespace argocd argo
helm ls | grep homelab | awk '{print($1)}' | xargs helm delete
kubectl delete -f services/auth.yaml -n homelab

