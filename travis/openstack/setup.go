package rax

import (
	"os"

	"github.com/metral/goutils"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/identity/v2/tokens"
)

func IdentitySetup() *tokens.Token {
	/*
		Depends on following ENV vars:
			OS_AUTH_URL
			OS_USERNAME
			OS_PASSWORD
			OS_TENANT_ID
	*/

	authOpts, err := openstack.AuthOptionsFromEnv()
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	provider, err := openstack.AuthenticatedClient(authOpts)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	client := openstack.NewIdentityV2(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})

	opts := tokens.AuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		Username:         os.Getenv("OS_USERNAME"),
		Password:         os.Getenv("OS_PASSWORD"),
		TenantID:         os.Getenv("OS_TENANT_ID"),
	}

	token, err := tokens.Create(client, opts).Extract()
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	return token
}
