kubectl create namespace argo
kubectl create namespace prod

helm install argo --namespace argo --values ./helm/values.yaml --create-namespace ./helm
kubectl apply -f helm/ci-workflow.yaml

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: git-creds
  namespace: argo
type: Opaque
stringData:
  username: ${GITHUB_USERNAME}
  password: ${GITHUB_TOKEN}
EOF

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: git-creds
  namespace: prod
type: Opaque
stringData:
  username: ${GITHUB_USERNAME}
  password: ${GITHUB_TOKEN}
EOF

IP=$(kubectl get svc -n argo | grep argo-docker-registry | awk '{print($3)}'):5000
sh ./build-pr-poller.sh $IP

ansible-playbook -i ansible/hosts ansible/run.yaml -e "master=$MASTER" -e "worker_1=$WORKER_1" -e "worker_2=$WORKER_2" -e "worker_3=$WORKER_3" -e "worker_4=$WORKER_4" -e "registry=$IP" --tags "configure_k3s"

kubectl apply -f helm/app.yaml
