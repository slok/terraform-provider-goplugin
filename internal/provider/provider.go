package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
	storagegit "github.com/slok/terraform-provider-goplugin/internal/plugin/storage/git"
	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	"github.com/slok/terraform-provider-goplugin/internal/provider/attributeutils"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func New() provider.Provider {
	return &tfProvider{}
}

type tfProvider struct {
	configured bool

	resourcePluginsV1   map[string]apiv1.ResourcePlugin
	dataSourcePluginsV1 map[string]apiv1.DataSourcePlugin
}

var pluginsAttributes = tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
	"source_code": {
		Required:    true,
		Description: `Configuration regarding where the plugin code will be loaded from. Only one must be used of all the methods available`,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"data": {
				Optional:    true,
				Description: `Raw content data of the plugins.`,
				Type:        types.ListType{ElemType: types.StringType},
			},
			"git": {
				Optional:    true,
				Description: `Git repository to get the plugin source data from.`,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						Required:    true,
						Description: `URL of the repository.`,
						Validators:  []tfsdk.AttributeValidator{attributeutils.NonEmptyString},
						Type:        types.StringType,
					},
					"ref": {
						Optional:      true,
						Description:   `Reference of the the repository, only Branch and tags are supported.`,
						PlanModifiers: tfsdk.AttributePlanModifiers{attributeutils.DefaultValue(types.String{Value: "main"})},
						Type:          types.StringType,
					},
					"paths_regex": {
						Required:    true,
						Description: `List of regex that will match the files that will be loaded as the plugin source data.`,
						Type:        types.ListType{ElemType: types.StringType},
					},
				}),
			},
		}),
	},

	"configuration": {
		Required:    true,
		Sensitive:   true,
		Description: `A JSON string object with the properties that will be passed to the plugin creation/initialization, the plugin is responsible of knowing how to load and use these properties (e.g: API tokens).`,
		Validators:  []tfsdk.AttributeValidator{attributeutils.NonEmptyString, attributeutils.MustJSONObject},
		Type:        types.StringType,
	},
})

// GetSchema returns the schema that the user must configure on the provider block.
func (p *tfProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
A Terraform provider to create terraform providers 🤯, but easier and faster!

Terraform go plugin provider is a Terraform provider that will let you execute Go plugins (using [yaegi](https://github.com/traefik/yaegi)) in terraform by implementing a very simple and small Go API.

- Check all full documentation on the repository [readme](https://github.com/slok/terraform-provider-goplugin).
- Check the [examples](https://github.com/slok/terraform-provider-goplugin/tree/main/examples) to see how to develop your own plugins.
- Check [Go v1 lib](https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1).
`,
		Attributes: map[string]tfsdk.Attribute{
			"resource_plugins_v1": {
				Optional:    true,
				Description: `The Block of resource plugins using v1 API that will be loaded by the provider.`,
				Attributes:  pluginsAttributes,
			},
			"data_source_plugins_v1": {
				Optional:    true,
				Description: `The Block of data source plugins using v1 API that will be loaded by the provider.`,
				Attributes:  pluginsAttributes,
			},
		},
	}, nil
}

// Provider configuration.
type providerData struct {
	ResourcePluginsV1   map[string]providerDataPluginV1 `tfsdk:"resource_plugins_v1"`
	DataSourcePluginsV1 map[string]providerDataPluginV1 `tfsdk:"data_source_plugins_v1"`
}

type providerDataPluginV1 struct {
	SourceCode    providerDataPluginV1Source `tfsdk:"source_code"`
	Configuration types.String               `tfsdk:"configuration"`
}

type providerDataPluginV1Source struct {
	Data []types.String                 `tfsdk:"data"`
	Git  *providerDataPluginV1SourceGit `tfsdk:"git"`
}
type providerDataPluginV1SourceGit struct {
	URL        types.String   `tfsdk:"url"`
	Ref        types.String   `tfsdk:"ref"`
	PathsRegex []types.String `tfsdk:"paths_regex"`
}

// This is like if it was our main entrypoint.
func (p *tfProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration.
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	v1factory := pluginv1.NewFactory()

	// Load resource plugins.
	resourcePlugins := map[string]apiv1.ResourcePlugin{}
	for pluginID, pluginConfig := range config.ResourcePluginsV1 {
		plugin, err := p.loadAPIV1ResourcePlugin(ctx, v1factory, pluginConfig)
		if err != nil {
			resp.Diagnostics.AddError("Error while loading resource plugin", fmt.Sprintf("Could not load plugin resource %q due to an error: %s", pluginID, err.Error()))
			return
		}
		resourcePlugins[pluginID] = plugin
	}

	// Load data source plugins.
	dataSourcePlugins := map[string]apiv1.DataSourcePlugin{}
	for pluginID, pluginConfig := range config.DataSourcePluginsV1 {
		plugin, err := p.loadAPIV1DataSourcePlugin(ctx, v1factory, pluginConfig)
		if err != nil {
			resp.Diagnostics.AddError("Error while loading data source plugin", fmt.Sprintf("Could not load data source plugin %q due to an error: %s", pluginID, err.Error()))
			return
		}
		dataSourcePlugins[pluginID] = plugin
	}

	p.configured = true
	p.resourcePluginsV1 = resourcePlugins
	p.dataSourcePluginsV1 = dataSourcePlugins
}

func (p *tfProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"goplugin_plugin_v1": resourcePluginV1Type{},
	}, nil
}

func (p *tfProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{
		"goplugin_plugin_v1": dataSourcePluginV1Type{},
	}, nil
}

func (p *tfProvider) loadAPIV1ResourcePlugin(ctx context.Context, pluginFactory *pluginv1.Factory, pluginConfig providerDataPluginV1) (apiv1.ResourcePlugin, error) {
	repo, err := p.loadAPIV1PluginSourceCode(ctx, pluginConfig.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin source code: %w", err)
	}

	plugin, err := pluginFactory.NewResourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    "NewResourcePlugin", // TODO(slok): Make it configurable by the user.
		PluginOptions:        pluginConfig.Configuration.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading plugin: %w", err)
	}

	return plugin, nil
}

func (p *tfProvider) loadAPIV1DataSourcePlugin(ctx context.Context, pluginFactory *pluginv1.Factory, pluginConfig providerDataPluginV1) (apiv1.DataSourcePlugin, error) {
	repo, err := p.loadAPIV1PluginSourceCode(ctx, pluginConfig.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin source code: %w", err)
	}

	plugin, err := pluginFactory.NewDataSourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    "NewDataSourcePlugin", // TODO(slok): Make it configurable by the user.
		PluginOptions:        pluginConfig.Configuration.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading plugin from source code: %w", err)
	}

	return plugin, nil
}

func (p *tfProvider) loadAPIV1PluginSourceCode(ctx context.Context, pluginConfig providerDataPluginV1Source) (storage.SourceCodeRepository, error) {
	// Select the source repo based on the configuration.
	switch {
	// Source code from raw data.
	case len(pluginConfig.Data) > 0:
		src := []string{}
		for _, s := range pluginConfig.Data {
			src = append(src, s.Value)
		}
		return storage.StaticSourceCodeRepository(src), nil

	// Source code from Git repository
	case pluginConfig.Git != nil:
		rgs := []*regexp.Regexp{}
		for _, rg := range pluginConfig.Git.PathsRegex {
			r, err := regexp.Compile(rg.Value)
			if err != nil {
				return nil, fmt.Errorf("could not compile %q regex: %w", rg, err)
			}

			rgs = append(rgs, r)
		}

		gitRepo, err := storagegit.NewSourceCodeRepository(storagegit.SourceCodeRepositoryConfig{
			URL:          pluginConfig.Git.URL.Value,
			BranchOrTag:  pluginConfig.Git.Ref.Value,
			MatchRegexes: rgs,
		})
		if err != nil {
			return nil, fmt.Errorf("could not obtain source code from git repository: %w", err)
		}

		return gitRepo, nil
	}

	// Invalid source code repo.
	return nil, fmt.Errorf("plugin source code source missing")
}
