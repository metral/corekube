package rax

import (
	"github.com/metral/goutils"
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

	client := openstack.NewIdentityV2(provider)

	opts := tokens.WrapOptions(authOpts)
	token, err := tokens.Create(client, opts).ExtractToken()
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})

	return token
}
