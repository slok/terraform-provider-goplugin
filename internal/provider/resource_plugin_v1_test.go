package provider_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

const providerConfigFmt = `
terraform {
  required_providers {
    goplugin = {
      source = "goplugin"
    }
  }
}

provider goplugin { 
  resource_plugins_v1 = {
    "test_file": {
      source_code = [file("testdata/file_plugin/plugin.go")]
      configuration =  jsonencode({})
    }
  }
}

%s
`

type expTestResourcePluginv1 struct {
	ID           string
	PluginID     string
	ResourceData string
}

func assertFileMissing(t *testing.T, file string) resource.TestCheckFunc {
	assert := assert.New(t)

	return resource.TestCheckFunc(func(s *terraform.State) error {
		_, err := os.Stat(file)
		assert.True(errors.Is(err, os.ErrNotExist), "File should not exist", file)

		return nil
	})
}

func assertFileExistsWithContent(t *testing.T, file string, expContent string) resource.TestCheckFunc {
	assert := assert.New(t)

	return resource.TestCheckFunc(func(s *terraform.State) error {
		data, err := os.ReadFile(file)
		assert.NoError(err)
		assert.Equal(expContent, string(data))

		return nil
	})
}

// TestAccResourcePlugingV1CreateDelete will check a plugin  handles create and delete correctly.
// This test relies on a test plugin that knows how to manage a file a terraform resource.
func TestAccResourcePlugingV1CreateDelete(t *testing.T) {
	tests := map[string]struct {
		config         string
		file           string
		expState       expTestResourcePluginv1
		expFile        string
		expFileContent string
		expErr         *regexp.Regexp
	}{
		"A missing plugin should fail.": {
			config: `
resource "goplugin_plugin_v1" "test" {
  plugin_id = "missing"
  resource_data = jsonencode({})
}
`,
			expErr: regexp.MustCompile(`"missing" plugin is not loaded`),
		},

		"A not json object 'resource_data' should fail.": {
			config: `
resource "goplugin_plugin_v1" "test" {
  plugin_id = "test_file"
  resource_data = "{"
}
`,
			expErr: regexp.MustCompile(`Attribute must be JSON object`),
		},

		"A correct configuration should execute correctly.": {
			config: `
resource "goplugin_plugin_v1" "test" {
  plugin_id = "test_file"
  resource_data = jsonencode({
    path = "/tmp/test.txt"
    content = "this is a test"
  })
}
`,
			expState: expTestResourcePluginv1{
				ID:           `test_file//tmp/test.txt`,
				PluginID:     `test_file`,
				ResourceData: `{"content":"this is a test","path":"/tmp/test.txt"}`,
			},
			expFile:        "/tmp/test.txt",
			expFileContent: "this is a test",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Prepare non error checks.
			var checks resource.TestCheckFunc
			if test.expErr == nil {
				checks = resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "id", test.expState.ID),
					resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "plugin_id", test.expState.PluginID),
					resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "resource_data", test.expState.ResourceData),
					assertFileExistsWithContent(t, test.expFile, test.expFileContent),
				)
			}

			// Assemble our Terraform config code.
			config := fmt.Sprintf(providerConfigFmt, test.config)

			// Execute test.
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             assertFileMissing(t, test.expFile),
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

// TestAccResourcePlugingV1Update will check a plugin  handles get and update correctly.
// This test relies on a test plugin that knows how to manage a file a terraform resource.
func TestAccResourcePlugingV1Update(t *testing.T) {
	tests := map[string]struct {
		configCreate         string
		configUpdate         string
		file                 string
		expState             expTestResourcePluginv1
		expFile              string
		expFileContentCreate string
		expFileContentUpdate string
	}{
		"A correct configuration should execute correctly.": {

			configCreate: `
resource "goplugin_plugin_v1" "test" {
  plugin_id = "test_file"
  resource_data = jsonencode({
    path = "/tmp/test.txt"
    content = "is this a test?"
  })
}
`,

			configUpdate: `
resource "goplugin_plugin_v1" "test" {
  plugin_id = "test_file"
  resource_data = jsonencode({
    path = "/tmp/test.txt"
    content = "this is a test"
  })
}
`,
			expState: expTestResourcePluginv1{
				ID:           `test_file//tmp/test.txt`,
				PluginID:     `test_file`,
				ResourceData: `{"content":"this is a test","path":"/tmp/test.txt"}`,
			},
			expFile:              "/tmp/test.txt",
			expFileContentCreate: "is this a test?",
			expFileContentUpdate: "this is a test",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// Assemble our Terraform config code.
			configCreate := fmt.Sprintf(providerConfigFmt, test.configCreate)
			configUpdate := fmt.Sprintf(providerConfigFmt, test.configUpdate)

			// Execute test.
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { testAccPreCheck(t) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				CheckDestroy:             assertFileMissing(t, test.expFile),
				Steps: []resource.TestStep{
					{
						ResourceName: "goplugin_plugin_v1.test",
						Config:       configCreate,
						Check: resource.ComposeAggregateTestCheckFunc(
							assertFileExistsWithContent(t, test.expFile, test.expFileContentCreate),
						),
					},
					{
						ResourceName:      "goplugin_plugin_v1.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
					{
						Config: configUpdate,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "id", test.expState.ID),
							resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "plugin_id", test.expState.PluginID),
							resource.TestCheckResourceAttr("goplugin_plugin_v1.test", "resource_data", test.expState.ResourceData),
							assertFileExistsWithContent(t, test.expFile, test.expFileContentUpdate),
						),
					},
				},
			})
		})
	}
}
