function run_ansible_task() {
	cd ansible
	ansible-playbook run.yaml -e "master=$MASTER" -e "worker_1=$WORKER_1" -e "worker_2=$WORKER_2" -e "worker_3=$WORKER_3" -e "worker_4=$WORKER_4" -e "registry=$IP" --tags $@
	cd ..
}

helm install argo --namespace argo --values ./helm/values.yaml --create-namespace ./helm
kubectl apply -f helm/app.yaml
kubectl apply -f helm/ci-workflow.yaml
kubectl apply -f helm/ci-secret.yaml

# configure K3s nodes for insecure docker registry
#IP=$(kubectl get svc | grep docker-registry | awk '{print($3)}')
#run_ansible_task "configure_k3s"
