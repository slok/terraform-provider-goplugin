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

func TestMustJSONObject(t *testing.T) {
	tests := map[string]struct {
		value  string
		expErr bool
	}{
		"Empty string should fail.": {
			value:  ``,
			expErr: true,
		},

		"Empty JSON object should not fail.": {
			value:  `{}`,
			expErr: false,
		},

		"Valid JSON object should not fail.": {
			value:  `{"x": "y"}`,
			expErr: false,
		},

		"JSON list should fail.": {
			value:  `   []`,
			expErr: true,
		},

		"Invalid JSON should fail.": {
			value:  `{`,
			expErr: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			val, err := types.StringType.ValueFromTerraform(context.TODO(), tftypes.NewValue(tftypes.String, test.value))
			require.NoError(err)

			request := tfsdk.ValidateAttributeRequest{
				AttributePath:   path.Root("test"),
				AttributeConfig: val,
			}
			response := &tfsdk.ValidateAttributeResponse{}

			attributeutils.MustJSONObject.Validate(context.TODO(), request, response)

			if test.expErr {
				assert.True(response.Diagnostics.HasError())
			} else {
				assert.False(response.Diagnostics.HasError())
			}
		})
	}
}
