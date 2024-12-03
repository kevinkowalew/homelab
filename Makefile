install:
	kubectl apply -f argo-app.yaml -n argocd
delete:
	kubectl delete -f argo-app.yaml -n argocd

