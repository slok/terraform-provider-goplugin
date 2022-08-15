package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func New() provider.Provider {
	return &tfProvider{}
}

type tfProvider struct {
	configured bool
}

// GetSchema returns the schema that the user must configure on the provider block.
func (p *tfProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
TODO.
`,
		Attributes: map[string]tfsdk.Attribute{},
	}, nil
}

// Provider configuration.
type providerData struct{}

// This is like if it was our main entrypoint.
func (p *tfProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration.
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.configured = true
}

func (p *tfProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{}, nil
}

func (p *tfProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}
