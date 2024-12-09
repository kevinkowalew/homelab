<img src="https://www.raspberrypi.com/app/uploads/2020/06/raspberrry_pi_logo.png" alt="raspberrypi" width="75"/> <img src="https://static-00.iconduck.com/assets.00/ansible-icon-512x512-fydu4n0b.png" alt="ansible" width="75"/> <img src="https://cdn4.iconfinder.com/data/icons/logos-and-brands/512/97_Docker_logo_logos-512.png" alt="docker" width="75"/> <img src="https://velog.velcdn.com/images/chane_ha_da/post/afdb6fd4-5619-4c94-8baf-5ac8e5d42369/image.webp" alt="kubernetes" width="75"/>  <img src="https://user-images.githubusercontent.com/686194/57031240-0cab6300-6bfc-11e9-9a24-b6806f41743f.png" alt="helm" width="75"/> <img src="https://argocd-image-updater.readthedocs.io/en/v0.10.0/assets/logo.png" alt="argo" width="75"/><img src="https://upload.wikimedia.org/wikipedia/commons/thumb/3/38/Prometheus_software_logo.svg/2066px-Prometheus_software_logo.svg.png" alt="prometheus" width="75"/> <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/Grafana_icon.svg/250px-Grafana_icon.svg.png" alt="grafana" width="75"/> <img src="https://w7.pngwing.com/pngs/448/730/png-transparent-postgresql-plain-logo-icon.png" alt="postgresl" width="75"/> <img src="https://images.ctfassets.net/o7xu9whrs0u9/7Hff1xq2vCM8DFNwfN0Bm8/ca9a237976a46ec08e63f37b57a56178/graphite-logo.png" alt="graphite" width="75"/>

This repo contains everything used to provision and manage my Kubernetes homelab:
- Infrastructure as Code (IaC) for provisioning vanilla `raspberry pi`s into `k3s` control plane nodes
- `GitOps` CI/CD pipeline for managing `go` services
- Monitoring Stack for tracking cluster & service health

# TODO
- [ ] update CI template to take these parameters: `repo`, `username`, `version`
- [ ] move datastores & monitoring into a dedicated `helm` chart & namespace
- [ ] create a microservices
- [ ] figure out where a UI can fill in operational gaps to make the UX pleasant

- you can put the Pods behind Services and use Service DNS for communication. Calls to service-name allow Pods in the same namespace to communicate. Calls to service-name.namespace allow Pods in different namespaces to communicate.
