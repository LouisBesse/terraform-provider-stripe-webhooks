package main

import (
	"terraform-provider-stripe-webhooks/stripe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: stripe.Provider,
	})
}
