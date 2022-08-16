package v1_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func TestResourcePluginCreate(t *testing.T) {
	tests := map[string]struct {
		pluginPaths []string
		request     apiv1.CreateResourceRequest
		expResponse *apiv1.CreateResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginPaths: []string{
				"./testdata/plugin_noop/plugin.go",
			},
			expResponse: &apiv1.CreateResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginPaths: []string{
				"./testdata/plugin_error/plugin.go",
			},
			expErr: true,
		},

		"A correct plugin should return the correct result.": {
			pluginPaths: []string{
				"./testdata/plugin_ok/plugin.go",
				"./testdata/plugin_ok/resource.go",
			},
			request: apiv1.CreateResourceRequest{
				ResourceData: "this is a test",
			},
			expResponse: &apiv1.CreateResourceResponse{
				ID: "this is a test_test1",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pluginSource := []string{}
			for _, p := range test.pluginPaths {
				d, err := os.ReadFile(p)
				require.NoError(err)
				pluginSource = append(pluginSource, string(d))
			}

			// Create the plugin twice to check the plugin cache.
			f := pluginv1.NewFactory()
			_, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)
			p, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)

			resp, err := p.CreateResource(context.TODO(), test.request)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, resp)
			}
		})
	}
}

func TestResourcePluginRead(t *testing.T) {
	tests := map[string]struct {
		pluginPaths []string
		request     apiv1.ReadResourceRequest
		expResponse *apiv1.ReadResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginPaths: []string{
				"./testdata/plugin_noop/plugin.go",
			},
			expResponse: &apiv1.ReadResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginPaths: []string{
				"./testdata/plugin_error/plugin.go",
			},
			expErr: true,
		},

		"A correct plugin should return the correct result.": {
			pluginPaths: []string{
				"./testdata/plugin_ok/plugin.go",
				"./testdata/plugin_ok/resource.go",
			},
			request: apiv1.ReadResourceRequest{
				ID: "this is a test",
			},
			expResponse: &apiv1.ReadResourceResponse{
				ResourceData: "this is a test_test1",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pluginSource := []string{}
			for _, p := range test.pluginPaths {
				d, err := os.ReadFile(p)
				require.NoError(err)
				pluginSource = append(pluginSource, string(d))
			}

			// Create the plugin twice to check the plugin cache.
			f := pluginv1.NewFactory()
			_, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)
			p, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)

			resp, err := p.ReadResource(context.TODO(), test.request)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, resp)
			}
		})
	}
}

func TestResourcePluginUpdate(t *testing.T) {
	tests := map[string]struct {
		pluginPaths []string
		request     apiv1.UpdateResourceRequest
		expResponse *apiv1.UpdateResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginPaths: []string{
				"./testdata/plugin_noop/plugin.go",
			},
			expResponse: &apiv1.UpdateResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginPaths: []string{
				"./testdata/plugin_error/plugin.go",
			},
			expErr: true,
		},

		"A correct plugin should return the correct result.": {
			pluginPaths: []string{
				"./testdata/plugin_ok/plugin.go",
				"./testdata/plugin_ok/resource.go",
			},
			request: apiv1.UpdateResourceRequest{
				ID:                "test1",
				ResourceData:      "This is",
				ResourceDataState: " a test",
			},
			expResponse: &apiv1.UpdateResourceResponse{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pluginSource := []string{}
			for _, p := range test.pluginPaths {
				d, err := os.ReadFile(p)
				require.NoError(err)
				pluginSource = append(pluginSource, string(d))
			}

			// Create the plugin twice to check the plugin cache.
			f := pluginv1.NewFactory()
			_, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)
			p, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)

			resp, err := p.UpdateResource(context.TODO(), test.request)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, resp)
			}
		})
	}
}

func TestResourcePluginDelete(t *testing.T) {
	tests := map[string]struct {
		pluginPaths []string
		request     apiv1.DeleteResourceRequest
		expResponse *apiv1.DeleteResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginPaths: []string{
				"./testdata/plugin_noop/plugin.go",
			},
			expResponse: &apiv1.DeleteResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginPaths: []string{
				"./testdata/plugin_error/plugin.go",
			},
			expErr: true,
		},

		"A correct plugin should return the correct result.": {
			pluginPaths: []string{
				"./testdata/plugin_ok/plugin.go",
				"./testdata/plugin_ok/resource.go",
			},
			request: apiv1.DeleteResourceRequest{
				ID: "test1",
			},
			expResponse: &apiv1.DeleteResourceResponse{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pluginSource := []string{}
			for _, p := range test.pluginPaths {
				d, err := os.ReadFile(p)
				require.NoError(err)
				pluginSource = append(pluginSource, string(d))
			}

			// Create the plugin twice to check the plugin cache.
			f := pluginv1.NewFactory()
			_, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)
			p, err := f.NewResourcePlugin(context.TODO(), pluginSource, "")
			require.NoError(err)

			resp, err := p.DeleteResource(context.TODO(), test.request)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, resp)
			}
		})
	}
}
