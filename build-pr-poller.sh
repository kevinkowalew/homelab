while true; do
  STATUS=$(kubectl -n argo get pods | grep argo-docker-registry | awk '{print($3)}')
  if [[ "$STATUS" == "Running" ]]; then
    break
  fi

  sleep 2
done

kubectl port-forward -n argo svc/argo-argo-workflows-server 2746:2746 > /dev/null &
PF_PID=$!

for i in {1..10}; do
  if nc -z localhost 2746; then
    break
  fi
  sleep 1
done

curl -X POST -H "Content-Type: application/json" localhost:2746/api/v1/workflows/argo --data-binary @- <<EOF
{
  "workflow": {
    "apiVersion": "argoproj.io/v1alpha1",
    "kind": "workflow",
    "metadata": {
      "name": "kevinkowalew-argo-pr-poller-v0.0.12",
      "namespace": "argo",
      "labels": {
        "workflows.argoproj.io/workflow-template": "github-ci-template"
      }
    },
    "spec": {
      "arguments": {
        "parameters": [
          {
            "name": "repo",
            "value": "kevinkowalew/argo-pr-poller"
          },
          {
            "name": "version",
            "value": "v0.0.12"
          },
          {
            "name": "registry",
            "value": "argo-docker-registry:5000"
          },
          {
            "name": "image",
            "value": "golang:1.23"
          },
          {
            "name": "revision",
            "value": "ad244a7638109395147adcbbc9146d979e2723e0"
          },
          {
            "name": "email",
            "value": "kowalewksi.ke@gmail.com"
          }
        ]
      },
      "workflowTemplateRef": {
        "name": "github-ci-template"
      }
    }
  }
}
EOF

while true; do
  PHASE=$(curl --silent http://localhost:2746/api/v1/workflows/argo | jq -r '.items[] | select(.metadata.name=="kevinkowalew-argo-pr-poller-v0.0.12") | .status.phase')

  if [[ "$PHASE" == "Succeeded" || "$PHASE" == "Failed" ]]; then
    break
  fi

  sleep 2
done


kill $PF_PID
wait $PF_PID 2>/dev/null
