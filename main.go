package main

import (
	"github.com/akaspin/terraform-provider-generic/generic"
	//"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	//"log"
)

var V string

func main() {
	//out, _ := logging.LogOutput()
	//log.SetOutput(out)

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return generic.Provider()
		},
	})
}
