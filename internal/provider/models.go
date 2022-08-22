package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourcePluginV1 struct {
	ID         types.String `tfsdk:"id"`
	ResourceID types.String `tfsdk:"resource_id"`
	PluginID   types.String `tfsdk:"plugin_id"`
	Attributes types.String `tfsdk:"attributes"`
}

type DataSourcePluginV1 struct {
	ID         types.String `tfsdk:"id"`
	PluginID   types.String `tfsdk:"plugin_id"`
	Attributes types.String `tfsdk:"attributes"`
	Result     types.String `tfsdk:"result"`
}
