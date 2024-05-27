package metrics

import (
	"github.com/Drumato/k8s-openstack-pv-exporter/openstack"
	"github.com/labstack/echo/v4"
)

// OndemandUpdateMetricsMiddleware はリクエスト受信時にオンデマンドでメトリクスを更新する
// 定期的にメトリクスを更新しておくアイデアもあるけど、それはOpenStack/Kubernetes APIコールが増えるので、
// リクエストを受け取ったときのみ更新するようにしておく
func OndemandUpdateMetricsMiddleware(client openstack.Client) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: メトリクスを更新する
			// updateMetrics()
			return next(c)
		}
	}
}
