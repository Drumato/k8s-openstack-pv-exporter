package openstack

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"

	gophercloud "github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	openstackgo "github.com/gophercloud/gophercloud/v2/openstack"
	volumesv3 "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

type Client interface {
	Config() ClientConfig
	ListVolumes(ctx context.Context, opts volumesv3.ListOptsBuilder) ([]volumesv3.Volume, error)
}

func NewDefaultClient(
	ctx context.Context,
	config ClientConfig,
) (Client, error) {

	pc, err := authenticate(
		ctx,
		config.AuthOptions,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dc := &DefaultClient{}

	bsClient, err := openstackgo.NewBlockStorageV3(pc, gophercloud.EndpointOpts{
		Region: config.RegionName,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dc.BlockStorageV3 = bsClient

	return dc, nil
}

type DefaultClient struct {
	BlockStorageV3 *gophercloud.ServiceClient
	config         ClientConfig
}

func (dc *DefaultClient) ListVolumes(ctx context.Context, opts volumesv3.ListOptsBuilder) ([]volumesv3.Volume, error) {
	page, err := volumesv3.List(dc.BlockStorageV3, opts).AllPages(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	volumes, err := volumesv3.ExtractVolumes(page)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return volumes, nil
}

func (dc *DefaultClient) Config() ClientConfig {
	return dc.config
}

func authenticate(
	ctx context.Context,
	authOptions gophercloud.AuthOptions,
) (*gophercloud.ProviderClient, error) {
	client, err := openstackgo.NewClient(authOptions.IdentityEndpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	client.UserAgent.Prepend("k8s-openstack-pv-exporter")

	cert := os.Getenv(CertEnvKey)
	ca := os.Getenv(CACertEnvKey)
	key := os.Getenv(KeyEnvKey)

	tlsConfig := &tls.Config{}
	if ca != "" {
		caPool := x509.NewCertPool()

		severCert, err := os.ReadFile(ca)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		caPool.AppendCertsFromPEM(severCert)

		tlsConfig.RootCAs = caPool
	}

	if cert != "" && key != "" {
		clientCert, err := os.ReadFile(cert)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		clientKey, err := os.ReadFile(key)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

	}

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}
	client.HTTPClient.Transport = transport

	if err := openstack.Authenticate(ctx, client, authOptions); err != nil {
		return nil, errors.WithStack(err)
	}

	return client, nil
}
