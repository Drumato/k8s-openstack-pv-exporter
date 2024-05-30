package metrics

import (
	"log/slog"

	"github.com/Drumato/k8s-openstack-pv-exporter/kubernetes"
	"github.com/Drumato/k8s-openstack-pv-exporter/openstack"
	"github.com/cockroachdb/errors"
	volumesv3 "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/labstack/echo/v4"
)

// OndemandUpdateMetricsMiddleware updates metrics on demand when a request is received.
// Although updating metrics periodically is an idea, it would increase OpenStack/Kubernetes API calls.
// Therefore, we update the metrics only when a request is received.
func OndemandUpdateMetricsMiddleware(
	logger *slog.Logger,
	openstackClient openstack.Client,
	k8sClient kubernetes.Client,
) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger.InfoContext(c.Request().Context(), "trying to list persistent volumes")
			pvs, err := k8sClient.ListPersistentVolumes(c.Request().Context())
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "failed to list persistent volumes", "error", err)
				return errors.WithStack(err)
			}
			logger.InfoContext(c.Request().Context(), "succeed to list persistent volumes", "length", len(pvs.Items))

			// To avoid N+1 API calls, receive the data as a list and match it accordingly.
			logger.InfoContext(c.Request().Context(), "trying to list openstack volumes")
			listOpts := volumesv3.ListOpts{
				AllTenants: false,
			}
			volumes, err := openstackClient.ListVolumes(
				c.Request().Context(),
				listOpts,
			)
			if err != nil {
				logger.ErrorContext(c.Request().Context(), "failed to list openstack volumes", "error", err)
				return errors.WithStack(err)
			}
			logger.InfoContext(c.Request().Context(), "succeed to list openstack volumes", "length", len(volumes))

			// TODO: O(M*N)
			for _, pv := range pvs.Items {
				logger.DebugContext(c.Request().Context(), "persistentvolume", "name", pv.Name, "namespace", pv.Namespace)
				for _, v := range volumes {
					logger.DebugContext(c.Request().Context(), "volume", "name", v.Name, "status", v.Status)
					isNotTargetPV := pv.Name != v.Name
					if isNotTargetPV {
						continue
					}

					labels := OpenStackVolumeStatusLabels{
						Name:      v.Name,
						Namespace: pv.Namespace,
						ID:        v.ID,
						Status:    v.Status,
					}
					if pv.Spec.ClaimRef != nil {
						labels.ClaimRef = pv.Spec.ClaimRef.Name
					}

					UpdateOpenStackVolumeStatusMetrics(labels, 1)
					for _, status := range openstack.VolumeStatusCatalog {
						if v.Status == status {
							continue
						}

						labels.Status = status
						UpdateOpenStackVolumeStatusMetrics(labels, 0)
					}
				}
			}
			return next(c)
		}
	}
}
