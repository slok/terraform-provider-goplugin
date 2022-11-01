package v1_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/moduledir"
	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

var (
	pluginDirNoop  = "./testdata/plugin_noop"
	pluginDirError = "./testdata/plugin_error"
	pluginDirOk    = "./testdata/plugin_ok"
)

func TestResourcePluginCreate(t *testing.T) {
	tests := map[string]struct {
		pluginDir   string
		request     apiv1.CreateResourceRequest
		expResponse *apiv1.CreateResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginDir:   pluginDirNoop,
			expResponse: &apiv1.CreateResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginDir: pluginDirError,
			expErr:    true,
		},

		"A correct plugin should return the correct result.": {
			pluginDir: pluginDirOk,
			request: apiv1.CreateResourceRequest{
				Attributes: "this is a test",
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

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.pluginDir))
			require.NoError(err)

			// Create the plugin twice to check the plugin cache.
			config := pluginv1.PluginConfig{
				SourceCodeRepository: repo,
				PluginOptions:        "",
				PluginFactoryName:    "NewResourcePlugin",
			}
			engine := pluginv1.NewEngine()
			_, err = engine.NewResourcePlugin(context.TODO(), config)
			require.NoError(err)
			p, err := engine.NewResourcePlugin(context.TODO(), config)
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
		pluginDir   string
		request     apiv1.ReadResourceRequest
		expResponse *apiv1.ReadResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginDir:   pluginDirNoop,
			expResponse: &apiv1.ReadResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginDir: pluginDirError,
			expErr:    true,
		},

		"A correct plugin should return the correct result.": {
			pluginDir: pluginDirOk,
			request: apiv1.ReadResourceRequest{
				ID: "this is a test",
			},
			expResponse: &apiv1.ReadResourceResponse{
				Attributes: "this is a test_test1",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.pluginDir))
			require.NoError(err)

			// Create the plugin twice to check the plugin cache.
			config := pluginv1.PluginConfig{
				SourceCodeRepository: repo,
				PluginOptions:        "",
				PluginFactoryName:    "NewResourcePlugin",
			}
			engine := pluginv1.NewEngine()
			_, err = engine.NewResourcePlugin(context.TODO(), config)
			require.NoError(err)
			p, err := engine.NewResourcePlugin(context.TODO(), config)
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
		pluginDir   string
		request     apiv1.UpdateResourceRequest
		expResponse *apiv1.UpdateResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginDir:   pluginDirNoop,
			expResponse: &apiv1.UpdateResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginDir: pluginDirError,
			expErr:    true,
		},

		"A correct plugin should return the correct result.": {
			pluginDir: pluginDirOk,
			request: apiv1.UpdateResourceRequest{
				ID:              "test1",
				Attributes:      "This is",
				AttributesState: " a test",
			},
			expResponse: &apiv1.UpdateResourceResponse{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.pluginDir))
			require.NoError(err)

			// Create the plugin twice to check the plugin cache.
			config := pluginv1.PluginConfig{
				SourceCodeRepository: repo,
				PluginOptions:        "",
				PluginFactoryName:    "NewResourcePlugin",
			}
			engine := pluginv1.NewEngine()
			_, err = engine.NewResourcePlugin(context.TODO(), config)
			require.NoError(err)
			p, err := engine.NewResourcePlugin(context.TODO(), config)
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
		pluginDir   string
		request     apiv1.DeleteResourceRequest
		expResponse *apiv1.DeleteResourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginDir:   pluginDirNoop,
			expResponse: &apiv1.DeleteResourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginDir: pluginDirError,
			expErr:    true,
		},

		"A correct plugin should return the correct result.": {
			pluginDir: pluginDirOk,
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

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.pluginDir))
			require.NoError(err)

			// Create the plugin twice to check the plugin cache.
			config := pluginv1.PluginConfig{
				SourceCodeRepository: repo,
				PluginOptions:        "",
				PluginFactoryName:    "NewResourcePlugin",
			}
			engine := pluginv1.NewEngine()
			_, err = engine.NewResourcePlugin(context.TODO(), config)
			require.NoError(err)
			p, err := engine.NewResourcePlugin(context.TODO(), config)
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

func TestDataSourcePluginRead(t *testing.T) {
	tests := map[string]struct {
		pluginDir   string
		request     apiv1.ReadDataSourceRequest
		expResponse *apiv1.ReadDataSourceResponse
		expErr      bool
	}{
		"Noop plugin should end correctly being a NOOP.": {
			pluginDir:   pluginDirNoop,
			expResponse: &apiv1.ReadDataSourceResponse{},
		},

		"Error plugin should fail the execution.": {
			pluginDir: pluginDirError,
			expErr:    true,
		},

		"A correct plugin should return the correct result.": {
			pluginDir: pluginDirOk,
			request: apiv1.ReadDataSourceRequest{
				Attributes: "this is a test",
			},
			expResponse: &apiv1.ReadDataSourceResponse{
				Result: "this is a testfrom_data_source",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(*testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.pluginDir))
			require.NoError(err)

			// Create the plugin twice to check the plugin cache.
			config := pluginv1.PluginConfig{
				SourceCodeRepository: repo,
				PluginOptions:        "",
				PluginFactoryName:    "NewDataSourcePlugin",
			}
			engine := pluginv1.NewEngine()
			_, err = engine.NewDataSourcePlugin(context.TODO(), config)
			require.NoError(err)
			p, err := engine.NewDataSourcePlugin(context.TODO(), config)
			require.NoError(err)

			resp, err := p.ReadDataSource(context.TODO(), test.request)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, resp)
			}
		})
	}
}
