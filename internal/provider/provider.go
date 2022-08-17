package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func New() provider.Provider {
	return &tfProvider{}
}

type tfProvider struct {
	configured bool

	loadedPluginsV1 map[string]apiv1.ResourcePlugin
}

// GetSchema returns the schema that the user must configure on the provider block.
func (p *tfProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
The Go plugin provider is used to create dynamically Terraform providers using small go plugins (provider inception!).

It removes all the complexity and bootstrapping that has a regular Terraform provider development, however this comes with some
limitations. So depends on your use case this may be handy.

The plugins based providers, are similar to the regular Terraform providers but differ somehow:

- Not compiled: We load Go plugins at runtime using Yaegi.
- No third party tools allowed: Only standard golang library and this provider libraries.
- Configuration of the provider and resources are based on JSON strings: Very dynamic, flexible and Go and Terraform have first class support for marshal/unmarshaling easily.
- Simplified small API: Designed and implemented focusing on maintainability, easy development and lowering the need of a user understanding low level terraform concepts.

## Plugins

TODO

## When to use it

TODO.

## Terraform cloud

The provider is portable, it's compatible with terraform cloud workers out of the box.
`,
		Attributes: map[string]tfsdk.Attribute{
			"resource_plugins_v1": {
				Required:    true,
				Description: `The Block of resource plugins using v1 API that will be loaded by the provider.`,
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"source_code": {
						Required:    true,
						Description: `The Source code of the plugin, allows multiple file data but must be of the same package.`,
						Type:        types.ListType{ElemType: types.StringType},
					},

					"configuration": {
						Required:    true,
						Sensitive:   true,
						Description: `A JSON string object with the properties that will be passed to the plugin creation/initialization, the plugin is responsible of knowing how to load and use these properties (e.g: API tokens).`,
						Validators:  []tfsdk.AttributeValidator{attributeutils.NonEmptyString, attributeutils.MustJSONObject},
						Type:        types.StringType,
					},
				}),
			},
		},
	}, nil
}

// Provider configuration.
type providerData struct {
	ResourcePluginsV1 map[string]providerDataPluginV1 `tfsdk:"resource_plugins_v1"`
}

type providerDataPluginV1 struct {
	SourceCode    []types.String `tfsdk:"source_code"`
	Configuration string         `tfsdk:"configuration"`
}

// This is like if it was our main entrypoint.
func (p *tfProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration.
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare plugins.
	plugins := map[string]apiv1.ResourcePlugin{}
	v1factory := pluginv1.NewFactory()

	for pluginID, plugin := range config.ResourcePluginsV1 {
		src := []string{}
		for _, v := range plugin.SourceCode {
			src = append(src, v.Value)
		}
		plugin, err := v1factory.NewResourcePlugin(ctx, src, plugin.Configuration)
		if err != nil {
			resp.Diagnostics.AddError("Error loading plugin", fmt.Sprintf("Could not load plugin %q: %s", pluginID, err.Error()))
			return
		}
		plugins[pluginID] = plugin
	}

	p.configured = true
	p.loadedPluginsV1 = plugins
}

func (p *tfProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"goplugin_plugin_v1": resourcePluginV1Type{},
	}, nil
}

func (p *tfProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}
