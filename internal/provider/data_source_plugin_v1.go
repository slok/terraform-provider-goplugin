package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
	v1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func newDataSourcePluginV1() datasource.DataSource {
	return &dataSourcePluginV1{}
}

type dataSourcePluginV1 struct {
	plugins map[string]apiv1.DataSourcePlugin
}

func (d *dataSourcePluginV1) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"attributes": {
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

func (d *dataSourcePluginV1) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin_v1"
}

func (d *dataSourcePluginV1) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	dsd, ok := req.ProviderData.(providerInstancedDataSourceData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected providerInstancedResourceData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.plugins = dsd.plugins
}

func (d *dataSourcePluginV1) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config.
	var tfConfig DataSourcePluginV1
	diags := req.Config.Get(ctx, &tfConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin.
	plugin, ok := d.plugins[tfConfig.PluginID.ValueString()]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", tfConfig.PluginID.ValueString()))
		return
	}

	// Execute plugin.
	pluginResp, err := plugin.ReadDataSource(ctx, v1.ReadDataSourceRequest{
		Attributes: tfConfig.Attributes.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	// Map result.
	tfConfig.Result = types.StringValue(pluginResp.Result)

	// Force execution every time.
	tfConfig.ID = types.StringValue(strconv.Itoa(int(time.Now().UnixNano())))

	diags = resp.State.Set(ctx, tfConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
