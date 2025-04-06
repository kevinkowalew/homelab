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

IP=$(kubectl get svc -n argo | grep argo-docker-registry | awk '{print($3)}'):5000
sh ./build-pr-poller.sh $IP

ansible-playbook -i ansible/hosts ansible/run.yaml -e "master=$MASTER" -e "worker_1=$WORKER_1" -e "worker_2=$WORKER_2" -e "worker_3=$WORKER_3" -e "worker_4=$WORKER_4" -e "registry=$IP" --tags "configure_k3s"

kubectl apply -f helm/app.yaml
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: argo-pr-poller
  namespace: argo
data:
  LOG_VERBOSITY: "INFO"
  REPO_LIST: "pr-tag-purger,argo-pr-poller,apigateway,infraapi"
  ARGO_HOST: "http://argo-argo-workflows-server:2746"
  ARGO_CI_EMAIL: "kowalewski.ke@gmail.com"
  ARGO_CI_TRIGGER: "#release"
  REGISTRY_HOST: "http://argo-docker-registry:5000"
  GITHUB_STATUS_CHECK_CI_HOST: "https://localhost:2746"
  VERSION_LOOKBACK: "0"
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: argo-pr-poller
  namespace: argo
  labels:
    app: argo-pr-poller
spec:
  schedule: "0/5 * * * *"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: argo-pr-poller
        spec:
          containers:
          - name: argo-pr-poller
            image:  ${IP}/kevinkowalew/argo-pr-poller:v0.0.13
            envFrom:
            - configMapRef:
                name: argo-pr-poller
            env:
            - name: GITHUB_USERNAME
              valueFrom:
                secretKeyRef:
                  name: git-creds
                  key: username
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: git-creds
                  key: password
            volumeMounts:
            - name: git-creds-volume
              mountPath: /mnt/git-creds
              readOnly: true
          volumes:
          - name: git-creds-volume
            secret:
              secretName: git-creds
          restartPolicy: Never
EOF
