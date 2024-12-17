function run_ansible_task() {
	cd ansible
	ansible-playbook run.yaml -e "master=$MASTER" -e "worker_1=$WORKER_1" -e "worker_2=$WORKER_2" -e "worker_3=$WORKER_3" -e "worker_4=$WORKER_4" -e "registry=$IP" --tags $@
	cd ..
}

# install ArgoCD
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl apply -f cd/app.yaml -n argocd

# install ArgoCI
kubectl create namespace argo --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.4/install.yaml
kubectl patch deployment argo-server -n argo --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/args", "value": [ "server", "--auth-mode=server" ]}]'
kubectl apply -f ci/ci.yaml

# install monitoring
kubectl create namespace homelab --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install homelab helm --values=helm/values.yaml -n homelab

# configure K3s nodes for insecure docker registry
IP=$(kubectl get svc | grep docker-registry | awk '{print($3)}')
run_ansible_task "configure_k3s"
