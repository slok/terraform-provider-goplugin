package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func newResourcePluginV1() resource.Resource {
	return &resourcePluginV1{}
}

type resourcePluginV1 struct {
	plugins map[string]apiv1.ResourcePlugin
}

func (r *resourcePluginV1) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
Executes a Resource Go plugin v1.

The requirements for a plugin are:

- Written in Go.
- No external dependencies, only Go standard library.
- Implemented in a single package.
- Implement the [plugin v1 API](https://github.com/slok/terraform-provider-goplugin/tree/main/pkg/api/v1).

Check [examples](https://github.com/slok/terraform-provider-goplugin/tree/main/examples)
`,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description:   "The ID of the terraform resource, also used on the import resource actions, it's composed by other 2 attributes in a specific format: `{plugin_id}/{resource_id}`.",
				Type:          types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Computed:      true,
			},
			"resource_id": {
				Description: `The ID of the resource itself (outside terraform), it's the one returned from
							  the plugins (E.g the UUID of a user returned from an external API),
							  normally this ID can be combined with a Datasource so the datsource knows the
							  ID of the resource that needs to get the data from.`,
				Type:          types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.UseStateForUnknown()},
				Computed:      true,
			},
			"plugin_id": {
				Description: `The ID of the plugin to use, must be loaded and registered by the provider.
							    To avoid inconsistencies, if changed the resource will be recreated.`,
				Type:          types.StringType,
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			},
			"attributes": {
				Description: `A JSON string object with the properties that will be passed to the plugin
								resource, the plugin is responsible of knowing how to load and use these properties.`,
				Type:          types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{attributeutils.SuppressEquivalentJSON},
				Validators:    []tfsdk.AttributeValidator{attributeutils.NonEmptyString, attributeutils.MustJSONObject},
				Required:      true,
			},
		},
	}, nil
}

func (r *resourcePluginV1) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin_v1"
}

func (r *resourcePluginV1) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	rd, ok := req.ProviderData.(providerInstancedResourceData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected providerInstancedResourceData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.plugins = rd.plugins
}

func (r *resourcePluginV1) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var tfResourcePlan ResourcePluginV1
	diags := req.Plan.Get(ctx, &tfResourcePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin.
	plugin, ok := r.plugins[tfResourcePlan.PluginID.ValueString()]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", tfResourcePlan.PluginID.ValueString()))
		return
	}

	// Execute plugin.
	pluginResp, err := plugin.CreateResource(ctx, apiv1.CreateResourceRequest{Attributes: tfResourcePlan.Attributes.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	if pluginResp.ID == "" {
		resp.Diagnostics.AddError("Plugin didn't return ID", fmt.Sprintf("On resource creation the plugin %q must return an ID, it didn't.", tfResourcePlan.PluginID.ValueString()))
		return
	}

	// Generate terraform ID.
	id := r.packID(tfResourcePlan.PluginID.ValueString(), pluginResp.ID)

	// Map result.
	newTfPluginV1 := ResourcePluginV1{
		ID:         types.StringValue(id),
		ResourceID: types.StringValue(pluginResp.ID),
		PluginID:   tfResourcePlan.PluginID,
		Attributes: tfResourcePlan.Attributes,
	}

	diags = resp.State.Set(ctx, newTfPluginV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourcePluginV1) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var tfResourceState ResourcePluginV1
	diags := req.State.Get(ctx, &tfResourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unpack ID.
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.ValueString(), err))
		return
	}

	// Get plugin.
	plugin, ok := r.plugins[pluginID]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", pluginID))
		return
	}

	// Execute plugin.
	pluginResp, err := plugin.ReadResource(ctx, apiv1.ReadResourceRequest{ID: resourceID})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	// Map result.
	newTfPluginV1 := ResourcePluginV1{
		ID:         tfResourceState.ID,
		ResourceID: types.StringValue(resourceID),
		PluginID:   types.StringValue(pluginID),
		Attributes: types.StringValue(pluginResp.Attributes),
	}

	diags = resp.State.Set(ctx, newTfPluginV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourcePluginV1) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve plan values.
	var tfResourcePlan ResourcePluginV1
	diags := req.Plan.Get(ctx, &tfResourcePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve state values.
	var tfResourceState ResourcePluginV1
	diags = req.State.Get(ctx, &tfResourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unpack ID.
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.ValueString(), err))
		return
	}

	// Get plugin.
	plugin, ok := r.plugins[pluginID]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", pluginID))
		return
	}

	// Execute plugin.
	_, err = plugin.UpdateResource(ctx, apiv1.UpdateResourceRequest{
		ID:              resourceID,
		Attributes:      tfResourcePlan.Attributes.ValueString(),
		AttributesState: tfResourceState.Attributes.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	// Map result.
	newTfPluginV1 := ResourcePluginV1{
		ID:         tfResourceState.ID,         // Once on state, never changes.
		ResourceID: tfResourceState.ResourceID, // Once on state, never changes.
		PluginID:   tfResourcePlan.PluginID,
		Attributes: tfResourcePlan.Attributes,
	}

	diags = resp.State.Set(ctx, newTfPluginV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourcePluginV1) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var tfResourceState ResourcePluginV1
	diags := req.State.Get(ctx, &tfResourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unpack ID.
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.ValueString(), err))
		return
	}

	// Get plugin.
	plugin, ok := r.plugins[pluginID]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", pluginID))
		return
	}

	// Execute plugin.
	_, err = plugin.DeleteResource(ctx, apiv1.DeleteResourceRequest{ID: resourceID})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	// Remove resource from state.
	resp.State.RemoveResource(ctx)
}

func (r *resourcePluginV1) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourcePluginV1) packID(pluginID, resourceID string) string {
	return pluginID + "/" + resourceID
}

func (r *resourcePluginV1) unpackID(id string) (pluginID, resourceID string, err error) {
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid resource ID format: %s (expected <PLUGIN ID>/<RESOURCE ID>)", id)
	}

	return s[0], s[1], nil
}
