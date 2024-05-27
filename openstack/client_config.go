package openstack

import (
	"os"

	gophercloud "github.com/gophercloud/gophercloud/v2"
)

type ClientConfig struct {
	gophercloud.AuthOptions
	RegionName string
}

func NewConfig() ClientConfig {
	return ClientConfig{}
}

func NewConfigFromEnv() ClientConfig {
	authURL := os.Getenv(AuthURLEnvKey)
	username := os.Getenv(UserNameEnvKey)
	password := os.Getenv(PasswordEnvKey)
	domainName := os.Getenv(DomainNameEnvKey)
	projectName := os.Getenv(ProjectNameEnvKey)
	if projectName == "" {
		projectName = os.Getenv(TenantNameEnvKey)
	}
	if domainName == "" {
		domainName = "Default"
	}
	regionName := os.Getenv(RegionNameEnvKey)

	return ClientConfig{
		AuthOptions: gophercloud.AuthOptions{
			IdentityEndpoint: authURL,
			Username:         username,
			Password:         password,
			DomainName:       domainName,
			TenantName:       projectName,
		},
		RegionName: regionName,
	}
}
