function run_ansible_task() {
	ansible-playbook ansible/run.yaml -e "master=$MASTER" -e "worker_1=$WORKER_1" -e "worker_2=$WORKER_2" -e "worker_3=$WORKER_3" -e "worker_4=$WORKER_4" -e "registry=$IP" --tags $@
	cd ..
}

# setup control plane nodes
run_ansible_task "node_setup"

# install docker on master
run_ansible_task "install_docker"
