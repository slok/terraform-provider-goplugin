package attributeutils

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MustJSONObject is a validator that will validate that a string is JSON object (e.g `{"x": "y"}`).
const MustJSONObject = mustJSONObject(false)

type mustJSONObject bool

func (m mustJSONObject) Description(ctx context.Context) string         { return "" }
func (m mustJSONObject) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (m mustJSONObject) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var s types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &s)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if s.IsUnknown() || s.IsNull() {
		return
	}

	var o any
	if err := json.Unmarshal([]byte(s.ValueString()), &o); err != nil {
		resp.Diagnostics.AddError(req.AttributePath.String(), "Attribute must be JSON object")
		return
	}

	if reflect.TypeOf(o).Kind() == reflect.Slice {
		resp.Diagnostics.AddError(req.AttributePath.String(), "JSON lists are not allowed")
		return
	}
}
