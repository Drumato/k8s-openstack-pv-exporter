package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/Drumato/k8s-openstack-pv-exporter/metrics"
	"github.com/Drumato/k8s-openstack-pv-exporter/openstack"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	exporterPort uint16
)

func Execute(ctx context.Context) error {
	c := &cobra.Command{
		RunE:          runE,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	c.Flags().Uint16VarP(&exporterPort, "port", "p", 8080, "the prometheus exporter port")

	return c.ExecuteContext(ctx)
}

func runE(c *cobra.Command, args []string) error {
	e := echo.New()

	client, err := openstack.NewDefault(c.Context(), openstack.NewConfigFromEnv())
	if err != nil {
		return errors.WithStack(err)
	}

	e.Use(metrics.OndemandUpdateMetricsMiddleware(client))
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	doneCh := make(chan bool, 1)
	go startHTTPServer(e, doneCh)

	// c.Context() that is inherited the SIGINT signal context is passed from main.go.
	<-c.Context().Done()
	if err := e.Shutdown(c.Context()); err != nil {
		return errors.WithStack(err)
	}

	<-doneCh
	return nil
}

func startHTTPServer(e *echo.Echo, doneCh chan<- bool) {
	addr := net.JoinHostPort("", fmt.Sprintf("%d", exporterPort))
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal("shutting down the server")
	}
	doneCh <- true
}
