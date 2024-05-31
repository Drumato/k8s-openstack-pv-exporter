# Kubernetes Openstack PV Exporter

This is a Prometheus exporter that collects information about OpenStack volumes and Kuberntes PersistentVolumes,
and exports their status as an GaugeVec metric.

## Roadmap

- [x] Basic behavior
- [ ] Additional useful options
- [ ] Helm chart

## How to use

The exporter can be used both within and outside a Kubernetes cluster.
A Docker image is available via [GitHub Packages](https://github.com/Drumato?tab=packages&repo_name=k8s-openstack-pv-exporter),
so you can use it in a Kubernetes manifest (you should prepare an ImagePullSecret for pulling the image from ghcr.io).

You need to set the following environment variables:

- `OS_AUTH_URL`
- `OS_USERNAME`
- `OS_PASSWORD`
- `OS_DOMAIN_NAME`
- `OS_TENANT_NAME` (or `OS_PROJECT_NAME` for older OpenStack environments)
- `OS_REGION_NAME`
- If `OS_CERT`, `OS_CACERT`, and `OS_KEY` are specified, the exporter will try to use them for authenticating with the OpenStack API.

To install to a Kubernetes cluster, Please read [Helm Chart](#./helm/README.md) Documentation.

## Exposed Metric

- `k8s_openstack_pv_exporter_openstack_volume_status` ... `GaugeVec`
  - Labels
    - `name` ... the name of OpenStack Volume/Kubernetes PersistentVolume.
    - `namespace` ... the namespace of the Kubernetes PersistentVolume.
    - `id` ... the resource id of the OpenStack Volume.
    - `claimRef` ... filled if the PersistentVolume is claimed from the other.
    - `status` ... the status of the OpenStack Volume

