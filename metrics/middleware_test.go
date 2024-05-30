package metrics_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Drumato/k8s-openstack-pv-exporter/kubernetes"
	"github.com/Drumato/k8s-openstack-pv-exporter/metrics"
	"github.com/Drumato/k8s-openstack-pv-exporter/openstack"
	volumesv3 "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type TestVolumeData struct {
	Namespace string
	Name      string
	ID        string
	ClaimRef  *string
	Status    string
}

func TestOndemandUpdateMetricsMiddleware_OK(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testVolumes := map[string]TestVolumeData{
		"volume-id1": {
			Namespace: "ns1",
			Name:      "volume1",
			ID:        "volume-id1",
			ClaimRef:  ptr.To("volume-claim-1"),
			Status:    "in-use",
		},
		"volume-id2": {
			Namespace: "ns1",
			Name:      "volume2",
			ID:        "volume-id2",
			ClaimRef:  ptr.To("volume-claim-2"),
			Status:    "availbale",
		},
		"volume-id3": {
			Namespace: "ns3",
			Name:      "volume3",
			ID:        "volume-id3",
			Status:    "detaching",
		},
	}
	testK8sClient := newFakePVGetClient(testVolumes)
	testOpenStackClient := newFakeVolumeGetClient(testVolumes)

	e := echo.New()
	rec := httptest.NewRecorder()

	reg := metrics.InitializeMetrics()
	handler := metrics.OndemandUpdateMetricsMiddleware(
		testLogger,
		testOpenStackClient,
		testK8sClient,
	)(echo.WrapHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)

	promDataParser := new(expfmt.TextParser)
	data, err := promDataParser.TextToMetricFamilies(rec.Body)
	assert.NoError(t, err)

	gauge := data["k8s_openstack_pv_exporter_openstack_volume_status"]
	for _, metric := range gauge.GetMetric() {
		labels := metric.GetLabel()
		labelMap := map[string]string{}
		for _, label := range labels {
			labelMap[label.GetName()] = label.GetValue()
		}

		id, ok := labelMap["id"]
		assert.True(t, ok, "'k8s_openstack_pv_exporter_openstack_volume_status' metric must have the 'id' label")
		testVolume, ok := testVolumes[id]
		// assert.True(t, ok, "the given id didn't be returned by OpenStack (fake) API")

		actualStatus, ok := labelMap["status"]
		if actualStatus == testVolume.Status {
			assert.Equal(t, float64(1), metric.GetGauge().GetValue(), "the gauge value must be 1 if the corresponding volume's status is same as the given volume")
		} else {
			assert.Equal(t, float64(0), metric.GetGauge().GetValue(), "the gauge value must be 0 if the corresponding volume's status is different from the given volume")
		}
	}
}

type fakePVGetClient struct {
	pvs map[string]TestVolumeData
}

func (c *fakePVGetClient) ListPersistentVolumes(ctx context.Context) (*corev1.PersistentVolumeList, error) {
	pvs := make([]corev1.PersistentVolume, len(c.pvs))

	for i := range c.pvs {
		pv :=
			corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:      c.pvs[i].Name,
					Namespace: c.pvs[i].Namespace,
				},
			}
		if c.pvs[i].ClaimRef != nil {
			pv.Spec.ClaimRef = &corev1.ObjectReference{
				Name: *c.pvs[i].ClaimRef,
			}
		}

		pvs = append(pvs, pv)
	}
	list := &corev1.PersistentVolumeList{
		Items: pvs,
	}
	return list, nil
}

func newFakePVGetClient(pvs map[string]TestVolumeData) kubernetes.Client {
	return &fakePVGetClient{pvs}
}

type fakeVolumeGetClient struct {
	volumes map[string]TestVolumeData
}

func (c *fakeVolumeGetClient) Config() openstack.ClientConfig {
	return openstack.ClientConfig{}
}

func newFakeVolumeGetClient(volumes map[string]TestVolumeData) openstack.Client {
	return &fakeVolumeGetClient{volumes}
}

func (c *fakeVolumeGetClient) ListVolumes(ctx context.Context, opts volumesv3.ListOptsBuilder) ([]volumesv3.Volume, error) {
	volumes := make([]volumesv3.Volume, len(c.volumes))

	for i := range c.volumes {
		volumes = append(volumes, volumesv3.Volume{
			ID:     c.volumes[i].ID,
			Name:   c.volumes[i].Name,
			Status: c.volumes[i].Status,
		})
	}

	return volumes, nil
}
