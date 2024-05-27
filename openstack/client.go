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
)

type Client interface {
}

func NewDefault(
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
		CA_Pool := x509.NewCertPool()

		severCert, err := os.ReadFile(ca)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		CA_Pool.AppendCertsFromPEM(severCert)

		tlsConfig.RootCAs = CA_Pool
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
