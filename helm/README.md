# Helm Chart Document

This helm chart deploys k8s-openstack-pv-exporter to your Kubernetes cluster.

## Install 

First of all, You need to deploy these resources before deploying this chart.

- A Secret for pulling the exporter image from `ghcr.io`
- [A Secret](#Secrets) that is used by the exporter for authenticating to the OpenStack API server.

```bash
helm repo add myrepo https://charts.mydomain.com/
helm install my-awesome-app myrepo/awesome-webapp
```

## Uninstall

``````bash
helm uninstall my-awesome-app
``````

## Secrets

### Auth Secret

the authentication Secret must follow the below format.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: 
type: kubernetes.io/tls
data:
  auth_url: "<OS_AUTH_URL>"
  username: "<OS_USERNAME>"
  password: "<OS_USERNAME>"
  domain_name: "<OS_USERNAME>"
  tenant_name: "<OS_TENAN>" # OS_PROJECT_NAME can also be used
  region_name: "<OS_USERNAME>"
  certificate: "<the contents of OS_CERT>"
  ca: "<the contents of OS_CACERT>"
  key: "<the contents of OS_KEY>"
```

