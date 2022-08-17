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

	loadedPluginsV1 map[string]apiv1.ResourcePlugin
}

// GetSchema returns the schema that the user must configure on the provider block.
func (p *tfProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
The Go plugin provider is used to create dynamically Terraform providers using small go plugins (provider inception!).

It removes all the complexity and bootstrapping that has a regular Terraform provider development, however this comes with some
limitations. So depends on your use case this may be handy.

The plugins based providers, are similar to the regular Terraform providers but differ somehow:

- Not compiled: We load Go plugins at runtime using Yaegi.
- No third party tools allowed: Only standard golang library and this provider libraries.
- Configuration of the provider and resources are based on JSON strings: Very dynamic, flexible and Go and Terraform have first class support for marshal/unmarshaling easily.
- Simplified small API: Designed and implemented focusing on maintainability, easy development and lowering the need of a user understanding low level terraform concepts.

## Plugins

TODO

## When to use it

TODO.

## Terraform cloud

The provider is portable, it's compatible with terraform cloud workers out of the box.
`,
		Attributes: map[string]tfsdk.Attribute{
			"resource_plugins_v1": {
				Required:    true,
				Description: `The Block of resource plugins using v1 API that will be loaded by the provider.`,
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
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
				}),
			},
		},
	}, nil
}

// Provider configuration.
type providerData struct {
	ResourcePluginsV1 map[string]providerDataResourcePluginV1 `tfsdk:"resource_plugins_v1"`
}

type providerDataResourcePluginV1 struct {
	SourceCode    providerDataResourcePluginV1Source `tfsdk:"source_code"`
	Configuration types.String                       `tfsdk:"configuration"`
}

type providerDataResourcePluginV1Source struct {
	Data []types.String                   `tfsdk:"data"`
	Git  *providerDataResourcePluginV1Git `tfsdk:"git"`
}
type providerDataResourcePluginV1Git struct {
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

	// Load plugins.
	plugins := map[string]apiv1.ResourcePlugin{}
	v1factory := pluginv1.NewFactory()
	for pluginID, pluginConfig := range config.ResourcePluginsV1 {
		plugin, err := p.loadAPIV1ResourcePlugin(ctx, v1factory, pluginConfig)
		if err != nil {
			resp.Diagnostics.AddError("Error while loading plugin", fmt.Sprintf("Could not load plugin %q due to an error: %s", pluginID, err.Error()))
			return
		}
		plugins[pluginID] = plugin
	}

	p.configured = true
	p.loadedPluginsV1 = plugins
}

func (p *tfProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"goplugin_plugin_v1": resourcePluginV1Type{},
	}, nil
}

func (p *tfProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
}

func (p *tfProvider) loadAPIV1ResourcePlugin(ctx context.Context, pluginFactory *pluginv1.Factory, pluginConfig providerDataResourcePluginV1) (apiv1.ResourcePlugin, error) {
	var repo storage.SourceCodeRepository

	// Select the source repo based on the configuration.
	switch {
	// Source code from raw data.
	case len(pluginConfig.SourceCode.Data) > 0:
		src := []string{}
		for _, s := range pluginConfig.SourceCode.Data {
			src = append(src, s.Value)
		}
		repo = storage.DataSourceCodeRepository(src)

	// Source code from Git respotiroy
	case pluginConfig.SourceCode.Git != nil:
		rgs := []*regexp.Regexp{}
		for _, rg := range pluginConfig.SourceCode.Git.PathsRegex {
			r, err := regexp.Compile(rg.Value)
			if err != nil {
				return nil, fmt.Errorf("could not compile %q regex: %w", rg, err)
			}

			rgs = append(rgs, r)
		}

		gitRepo, err := storagegit.NewSourceCodeRepository(storagegit.SourceCodeRepositoryConfig{
			URL:          pluginConfig.SourceCode.Git.URL.Value,
			BranchOrTag:  pluginConfig.SourceCode.Git.Ref.Value,
			MatchRegexes: rgs,
		})
		if err != nil {
			return nil, fmt.Errorf("could not obtain source code from git repository: %w", err)
		}
		repo = gitRepo

	// Invalid source code repo.
	default:
		return nil, fmt.Errorf("plugin source code source missing")
	}

	plugin, err := pluginFactory.NewResourcePlugin(ctx, repo, pluginConfig.Configuration.Value)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin from source code: %w", err)
	}

	return plugin, nil
}
