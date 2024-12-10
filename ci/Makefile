install-argo:
	kubectl create namespace argo
	kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.4.4/install.yaml
	kubectl patch deployment \
	  argo-server \
	  --namespace argo \
	  --type='json' \
	  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/args", "value": [
	  "server",
	  "--auth-mode=server"
	]}]'
install-ci-pipeline:
	k apply -f github-ci-template.yaml
delete-ci-pipeline:
	k delete -f github-ci-template.yaml


