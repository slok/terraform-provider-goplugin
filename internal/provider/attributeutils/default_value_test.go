package attributeutils_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
)

func TestDefaultValue(t *testing.T) {
	tests := map[string]struct {
		configValue attr.Value
		planValue   attr.Value
		defValue    attr.Value
		expValue    attr.Value
		expErr      bool
	}{
		"Value with config should not default.": {
			configValue: types.StringValue("This is a value"),
			planValue:   types.StringValue("This is a value"),
			defValue:    types.StringValue("This is a default value"),
			expValue:    types.StringValue("This is a value"),
		},

		"Null config should default.": {
			configValue: types.StringNull(),
			planValue:   types.StringNull(),
			defValue:    types.StringValue("This is a default value"),
			expValue:    types.StringValue("This is a default value"),
		},

		"Not null plan should not default.": {
			configValue: types.StringNull(),
			planValue:   types.StringValue("This is a value"),
			defValue:    types.StringValue("This is a default value"),
			expValue:    types.StringValue("This is a value"),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			request := tfsdk.ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributePlan:   test.planValue,
				AttributeConfig: test.configValue,
			}

			response := &tfsdk.ModifyAttributePlanResponse{
				AttributePlan: test.planValue,
			}

			attributeutils.DefaultValue(test.defValue).Modify(context.TODO(), request, response)

			if test.expErr {
				assert.True(response.Diagnostics.HasError())
			} else {
				assert.False(response.Diagnostics.HasError())
				assert.Equal(test.expValue, response.AttributePlan)
			}
		})
	}
}
