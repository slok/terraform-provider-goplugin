package attributeutils_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
)

func TestSuppressEquivalentJSON(t *testing.T) {
	tests := map[string]struct {
		jsonState  string
		jsonConfig string
		expErr     bool
		expEquals  bool
	}{
		"Empty JSON should be treated as equals.": {
			jsonState:  "{}",
			jsonConfig: "{}",
			expEquals:  true,
		},

		"Same string should be treated as equals.": {
			jsonState:  `{"x": "y"}`,
			jsonConfig: `{"x": "y"}`,
			expEquals:  true,
		},

		"spaces, newlines, tavs... should be treated as equals": {
			jsonState: `
{
	"x":"y"     ,
	"a": "b"
}`,
			jsonConfig: `{"x":"y", "a": "b"}`,
			expEquals:  true,
		},

		"Different order but same content should be treated as equals.": {
			jsonState:  `{"x": "y", "a": "b"}`,
			jsonConfig: `{"a": "b", "x": "y"}`,
			expEquals:  true,
		},

		"Different  content should be treated as different.": {
			jsonState:  `{"x": "y", "a": "b"}`,
			jsonConfig: `{"a": "b", "x": "z"}`,
			expEquals:  false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			jsonState, err := types.StringType.ValueFromTerraform(context.TODO(), tftypes.NewValue(tftypes.String, test.jsonState))
			require.NoError(err)
			jsonConfig, err := types.StringType.ValueFromTerraform(context.TODO(), tftypes.NewValue(tftypes.String, test.jsonConfig))
			require.NoError(err)

			request := tfsdk.ModifyAttributePlanRequest{
				AttributePath:   path.Root("test"),
				AttributeState:  jsonState,
				AttributeConfig: jsonConfig,
			}

			response := &tfsdk.ModifyAttributePlanResponse{
				AttributePlan: jsonConfig,
			}

			attributeutils.SuppressEquivalentJSON.Modify(context.TODO(), request, response)

			if test.expErr {
				assert.True(response.Diagnostics.HasError())
			} else {
				assert.False(response.Diagnostics.HasError())

				// IF expect equals, the response should have the state data.
				if test.expEquals {
					assert.Equal(jsonState, response.AttributePlan)
				} else {
					assert.Equal(jsonConfig, response.AttributePlan)
				}
			}
		})
	}
}
