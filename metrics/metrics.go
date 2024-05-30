package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OpenStackVolumeStatusGaugeVec *prometheus.GaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_openstack_pv_exporter",
		Name:      "openstack_volume_status",
		Help:      "the status of an OpenStack Volume that relates to a Kubernetes PersistentVolume",
	}, []string{
		"name",
		"namespace",
		"id",
		"claimRef",
		"status",
	})
)

// メトリクス更新時、どのようなラベルをつければいいのか更新側から分かりづらいので構造体経由で渡すようにする
type OpenStackVolumeStatusLabels struct {
	// Name is the name of an OpenStack volume.
	Name string
	// Namespace is a Kubernetes namespace.
	Namespace string
	// ID is the resource id of an OpenStack volume.
	ID string
	// ClaimRef is the related k8s resource that claims the persistentvolume.
	ClaimRef string
	// Status is the resource status of an OpenStack volume.
	Status string
}

func InitializeMetrics() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(OpenStackVolumeStatusGaugeVec)
	return reg
}

func UpdateOpenStackVolumeStatusMetrics(labels OpenStackVolumeStatusLabels, value float64) {
	OpenStackVolumeStatusGaugeVec.With(prometheus.Labels{
		"name":      labels.Name,
		"namespace": labels.Namespace,
		"id":        labels.ID,
		"claimRef":  labels.ClaimRef,
		"status":    labels.Status,
	}).Set(value)
}
