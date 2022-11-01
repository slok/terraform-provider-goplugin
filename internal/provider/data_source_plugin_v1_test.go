package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const dataSourceProviderConfigFmt = `
terraform {
  required_providers {
    goplugin = {
      source = "goplugin"
    }
  }
}

provider goplugin { 
  data_source_plugins_v1 = {
    "fake": {
      source_code = {
		dir = "testdata/fake_data_source"
	  }
      configuration =  jsonencode({})
    }
  }
}

%s
`

// TestAccDataSourcePlugingV1 will check a data source plugin is executed correctly.
func TestAccDataSourcePlugingV1(t *testing.T) {
	tests := map[string]struct {
		config    string
		expResult string
		expErr    *regexp.Regexp
	}{
		"A missing plugin should fail.": {
			config: `
data "goplugin_plugin_v1" "test" {
  plugin_id = "missing"
  attributes = jsonencode({})
}
`,
			expErr: regexp.MustCompile(`"missing" plugin is not loaded`),
		},

		"A not json object 'attributes' should fail.": {
			config: `
data "goplugin_plugin_v1" "test" {
  plugin_id = "fake"
  attributes = "{"
}
`,
			expErr: regexp.MustCompile(`Attribute must be JSON object`),
		},

		"A correct configuration should execute correctly.": {
			config: `
data "goplugin_plugin_v1" "test" {
  plugin_id = "fake"
  attributes = jsonencode({
    something = "otherthing"
    test1 = "test2"
  })
}
`,
			expResult: `{"something":"otherthing","test1":"test2"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Prepare non error checks.
			var checks resource.TestCheckFunc
			if test.expErr == nil {
				checks = resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.goplugin_plugin_v1.test", "result", test.expResult),
				)
			}

			// Assemble our Terraform config code.
			config := fmt.Sprintf(dataSourceProviderConfigFmt, test.config)

			// Execute test.
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      config,
						Check:       checks,
						ExpectError: test.expErr,
					},
				},
			})
		})
	}
}
