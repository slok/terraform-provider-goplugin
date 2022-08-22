package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type resourcePluginV1Type struct{}

func (r resourcePluginV1Type) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (r resourcePluginV1Type) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	prv := p.(*tfProvider)
	return resourcePluginV1{
		p: prv,
	}, nil
}

type resourcePluginV1 struct {
	p *tfProvider
}

func (r resourcePluginV1) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError("Provider not configured", "The provider hasn't been configured before apply.")
		return
	}

	// Retrieve values from plan.
	var tfResourcePlan ResourcePluginV1
	diags := req.Plan.Get(ctx, &tfResourcePlan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin.
	plugin, ok := r.p.resourcePluginsV1[tfResourcePlan.PluginID.Value]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", tfResourcePlan.PluginID.Value))
		return
	}

	// Execute plugin.
	pluginResp, err := plugin.CreateResource(ctx, apiv1.CreateResourceRequest{Attributes: tfResourcePlan.Attributes.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error executing plugin", "Plugin execution end in error: "+err.Error())
		return
	}

	if pluginResp.ID == "" {
		resp.Diagnostics.AddError("Plugin didn't return ID", fmt.Sprintf("On resource creation the plugin %q must return an ID, it didn't.", tfResourcePlan.PluginID.Value))
		return
	}

	// Generate terraform ID.
	id := r.packID(tfResourcePlan.PluginID.Value, pluginResp.ID)

	// Map result.
	newTfPluginV1 := ResourcePluginV1{
		ID:         types.String{Value: id},
		ResourceID: types.String{Value: pluginResp.ID},
		PluginID:   tfResourcePlan.PluginID,
		Attributes: tfResourcePlan.Attributes,
	}

	diags = resp.State.Set(ctx, newTfPluginV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourcePluginV1) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError("Provider not configured", "The provider hasn't been configured before apply.")
		return
	}

	// Retrieve values from state.
	var tfResourceState ResourcePluginV1
	diags := req.State.Get(ctx, &tfResourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unpack ID.
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.Value, err))
		return
	}

	// Get plugin.
	plugin, ok := r.p.resourcePluginsV1[pluginID]
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
		ResourceID: types.String{Value: resourceID},
		PluginID:   types.String{Value: pluginID},
		Attributes: types.String{Value: pluginResp.Attributes},
	}

	diags = resp.State.Set(ctx, newTfPluginV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourcePluginV1) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError("Provider not configured", "The provider hasn't been configured before apply.")
		return
	}

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
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.Value, err))
		return
	}

	// Get plugin.
	plugin, ok := r.p.resourcePluginsV1[pluginID]
	if !ok {
		resp.Diagnostics.AddError("Plugin missing", fmt.Sprintf("%q plugin is not loaded", pluginID))
		return
	}

	// Execute plugin.
	_, err = plugin.UpdateResource(ctx, apiv1.UpdateResourceRequest{
		ID:              resourceID,
		Attributes:      tfResourcePlan.Attributes.Value,
		AttributesState: tfResourceState.Attributes.Value,
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

func (r resourcePluginV1) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError("Provider not configured", "The provider hasn't been configured before apply.")
		return
	}

	// Retrieve values from state.
	var tfResourceState ResourcePluginV1
	diags := req.State.Get(ctx, &tfResourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unpack ID.
	pluginID, resourceID, err := r.unpackID(tfResourceState.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("ID is wrong", fmt.Sprintf("%q id is wrong: %s", tfResourceState.ID.Value, err))
		return
	}

	// Get plugin.
	plugin, ok := r.p.resourcePluginsV1[pluginID]
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

func (r resourcePluginV1) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r resourcePluginV1) packID(pluginID, resourceID string) string {
	return pluginID + "/" + resourceID
}

func (r resourcePluginV1) unpackID(id string) (pluginID, resourceID string, err error) {
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf(
			"invalid resource ID format: %s (expected <PLUGIN ID>/<RESOURCE ID>)", id)
	}

	return s[0], s[1], nil
}
