package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Drumato/k8s-openstack-pv-exporter/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}
