package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	v1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type dataSourcePluginV1Type struct{}

func (d dataSourcePluginV1Type) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Executes a Data source Go plugin v1.

The requirements for a plugin are:

- Written in Go.
- No external dependencies, only Go standard library.
- Implemented in a single package.
- Implement the [plugin v1 API](https://github.com/slok/terraform-provider-goplugin/tree/main/pkg/api/v1).

Check [examples](https://github.com/slok/terraform-provider-goplugin/tree/main/examples)
`,
		Attributes: map[string]tfsdk.Attribute{
			"plugin_id": {
				Description: `The ID of the data source plugin to use, must be loaded and registered by the provider.`,
				Type:        types.StringType,
				Required:    true,
			},
			"arguments": {
				Description: `A JSON string object with the properties that will be passed to the data source
							  plugin, the plugin is responsible of knowing how to load and use these properties.`,
				Type:       types.StringType,
				Validators: []tfsdk.AttributeValidator{attributeutils.NonEmptyString, attributeutils.MustJSONObject},
				Required:   true,
			},
			"result": {
				Description: `A JSON string object with the plugin result.`,
				Type:        types.StringType,
				Computed:    true,
			},
			"id": {
				Description: `Not used (Used internally by the provider and Terraform), can be ignored.`,
				Computed:    true,
				Type:        types.StringType,
			},
		},
	}, nil
}

func (d dataSourcePluginV1Type) NewDataSource(ctx context.Context, p provider.Provider) (datasource.DataSource, diag.Diagnostics) {
	prv := p.(*tfProvider)
	return dataSourcePluginV1{
		p: prv,
	}, nil
}

type dataSourcePluginV1 struct {
	p *tfProvider
}

func (d dataSourcePluginV1) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.configured {
		resp.Diagnostics.AddError("Provider not configured", "The provider hasn't been configured before apply.")
		return
	}

	// Retrieve values from config.
	var tfConfig DataSourcePluginV1
	diags := req.Config.Get(ctx, &tfConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin.
	plugin, ok := d.p.dataSourcePluginsV1[tfConfig.PluginID.Value]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", tfConfig.PluginID.Value))
		return
	}

	// Execute plugin.
	pluginResp, err := plugin.ReadDataSource(ctx, v1.ReadDataSourceRequest{
		Arguments: tfConfig.Arguments.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	// Map result.
	tfConfig.Result = types.String{Value: pluginResp.Result}

	// Force execution every time.
	tfConfig.ID = types.String{Value: strconv.Itoa(int(time.Now().UnixNano()))}

	diags = resp.State.Set(ctx, tfConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
