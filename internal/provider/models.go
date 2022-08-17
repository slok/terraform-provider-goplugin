package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourcePluginV1 struct {
	ID           types.String `tfsdk:"id"`
	PluginID     types.String `tfsdk:"plugin_id"`
	ResourceData types.String `tfsdk:"resource_data"`
}
