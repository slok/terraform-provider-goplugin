package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/slok/terraform-provider-goplugin/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"goplugin": providerserver.NewProtocol6WithError(provider.New()),
}

func testAccPreCheck(t *testing.T) {}
