package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
	storagegit "github.com/slok/terraform-provider-goplugin/internal/plugin/storage/git"
	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/moduledir"
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

var (
	pluginSourceCodeAttribute = tfsdk.Attribute{
		Required: true,
		Description: `Configuration regarding where the plugin code will be loaded from.
		The plugin must be a valid go module ` + "(`go.mod`)" + ` and be available in the root this module.
		Only one of the source code retrieval methods must be used.`,
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"dir": {
				Optional:    true,
				Description: `Directory where the plugin go module root is. It will load all files including vendor directory, factories must be at the module root level, however it can have subpacakges.`,
				Type:        types.StringType,
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
						Optional:    true,
						Description: `Reference of the the repository, only Branch and tags are supported.`,
						Type:        types.StringType,
						// TODO(slok): Provider config doesn't support plan modifiers, set default on `Configure` until are supported.
						// PlanModifiers: tfsdk.AttributePlanModifiers{attributeutils.DefaultValue(types.String{Value: "main"})},
					},
					"dir": {
						Optional:    true,
						Description: "Absolute directory from the root where the plugin go module is in the repository. It works the same way the `dir` source code does, supports subpackages, vendor dir...",
						Type:        types.StringType,
					},
					"auth": {
						Optional:    true,
						Description: `Optional git authentication, if block exists it will enable (and also try loading env vars), if missing all auth will be disabled (not loading env vars).`,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"username": {
								Optional:    true,
								Description: "The username of the basic auth, if not set it will fallback to `GOPLUGIN_GIT_USERNAME` env var (Note: Github PATs don't need username).",
								Type:        types.StringType,
							},
							"password": {
								Optional:    true,
								Description: "The password of the basic auth, if not set it will fallback to `GOPLUGIN_GIT_PASSWORD` env var (Note: Github PATs can be used as passwords).",
								Type:        types.StringType,
							},
						}),
					},
				}),
			},
		}),
	}

	pluginConfigurationAttribute = tfsdk.Attribute{
		Required:    true,
		Sensitive:   true,
		Description: `A JSON string object with the properties that will be passed to the plugin creation/initialization, the plugin is responsible of knowing how to load and use these properties (e.g: API tokens).`,
		Validators:  []tfsdk.AttributeValidator{attributeutils.NonEmptyString, attributeutils.MustJSONObject},
		Type:        types.StringType,
	}
)

// GetSchema returns the schema that the user must configure on the provider block.
func (p *tfProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
A Terraform provider to create terraform providers, but easier and faster!

Terraform go plugin provider is a Terraform provider that will let you execute Go plugins (using [yaegi](https://github.com/traefik/yaegi)) in terraform by implementing a very simple and small Go API.

- Check all full documentation on the repository [readme](https://github.com/slok/terraform-provider-goplugin).
- Check the [examples](https://github.com/slok/terraform-provider-goplugin/tree/main/examples) to see how to develop your own plugins.
- Check [Go v1 lib](https://pkg.go.dev/github.com/slok/terraform-provider-goplugin/pkg/api/v1).
`,
		Attributes: map[string]tfsdk.Attribute{
			"resource_plugins_v1": {
				Optional:    true,
				Description: `The Block of resource plugins using v1 API that will be loaded by the provider.`,
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"source_code":   pluginSourceCodeAttribute,
					"configuration": pluginConfigurationAttribute,
					"factory_name": {
						Optional: true,
						Description: "The name of the plugin factory (in the source code) that will be used to make instances of the plugin, `NewResourcePlugin` by default, " +
							"specially helpful when a package has multiple plugins inside the same package so it can reuse parts of the code between all the plugins.",
						Type: types.StringType,
						// TODO(slok): Provider config doesn't support plan modifiers, set default on `Configure` until are supported.
						// PlanModifiers: tfsdk.AttributePlanModifiers{attributeutils.DefaultValue(types.String{Value: "NewResourcePlugin"})},
					},
				}),
			},
			"data_source_plugins_v1": {
				Optional:    true,
				Description: `The Block of data source plugins using v1 API that will be loaded by the provider.`,
				Attributes: tfsdk.MapNestedAttributes(map[string]tfsdk.Attribute{
					"source_code":   pluginSourceCodeAttribute,
					"configuration": pluginConfigurationAttribute,
					"factory_name": {
						Optional: true,
						Description: "The name of the plugin factory (in the source code) that will be used to make instances of the plugin, `NewDataSourcePlugin` by default, " +
							"specially helpful when a package has multiple plugins inside the same package so it can reuse parts of the code between all the plugins.",
						Type: types.StringType,
						// TODO(slok): Provider config doesn't support plan modifiers, set default on `Configure` until are supported.
						// PlanModifiers: tfsdk.AttributePlanModifiers{attributeutils.DefaultValue(types.String{Value: "NewDataSourcePlugin"})},
					},
				}),
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
	FactoryName   types.String               `tfsdk:"factory_name"`
}

type providerDataPluginV1Source struct {
	Dir types.String                   `tfsdk:"dir"`
	Git *providerDataPluginV1SourceGit `tfsdk:"git"`
}
type providerDataPluginV1SourceGit struct {
	URL  types.String                       `tfsdk:"url"`
	Ref  types.String                       `tfsdk:"ref"`
	Auth *providerDataPluginV1SourceGitAuth `tfsdk:"auth"`
	Dir  types.String                       `tfsdk:"dir"`
}
type providerDataPluginV1SourceGitAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
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

	pluginV1Engine := pluginv1.NewEngine()

	// Load resource plugins.
	resourcePlugins := map[string]apiv1.ResourcePlugin{}
	for pluginID, pluginConfig := range config.ResourcePluginsV1 {
		plugin, err := p.loadAPIV1ResourcePlugin(ctx, pluginV1Engine, pluginConfig)
		if err != nil {
			resp.Diagnostics.AddError("Error while loading resource plugin", fmt.Sprintf("Could not load plugin resource %q due to an error: %s", pluginID, err.Error()))
			return
		}
		resourcePlugins[pluginID] = plugin
	}

	// Load data source plugins.
	dataSourcePlugins := map[string]apiv1.DataSourcePlugin{}
	for pluginID, pluginConfig := range config.DataSourcePluginsV1 {
		plugin, err := p.loadAPIV1DataSourcePlugin(ctx, pluginV1Engine, pluginConfig)
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

func (p *tfProvider) loadAPIV1ResourcePlugin(ctx context.Context, pluginFactory *pluginv1.Engine, pluginConfig providerDataPluginV1) (apiv1.ResourcePlugin, error) {
	repo, err := p.loadAPIV1PluginSourceCode(ctx, pluginConfig.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin source code: %w", err)
	}

	// TODO(slok): Remove when plan modifiers are supported on provider configuration attributes.
	factoryName := pluginConfig.FactoryName.Value
	if factoryName == "" {
		factoryName = apiv1.DefaultResourcePluginFactoryName
	}

	plugin, err := pluginFactory.NewResourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    factoryName,
		PluginOptions:        pluginConfig.Configuration.Value,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading plugin: %w", err)
	}

	return plugin, nil
}

func (p *tfProvider) loadAPIV1DataSourcePlugin(ctx context.Context, pluginFactory *pluginv1.Engine, pluginConfig providerDataPluginV1) (apiv1.DataSourcePlugin, error) {
	repo, err := p.loadAPIV1PluginSourceCode(ctx, pluginConfig.SourceCode)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin source code: %w", err)
	}

	// TODO(slok): Remove when plan modifiers are supported on provider configuration attributes.
	factoryName := pluginConfig.FactoryName.Value
	if factoryName == "" {
		factoryName = apiv1.DefaultDataSourcePluginFactoryName
	}

	plugin, err := pluginFactory.NewDataSourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    factoryName,
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
	// Source code from fs dir.
	case pluginConfig.Dir.Value != "":
		realFS := os.DirFS(pluginConfig.Dir.Value)
		return moduledir.NewSourceCodeRepository(realFS)

	// Source code from Git repository
	case pluginConfig.Git != nil:
		// TODO(slok): Remove when plan modifiers are supported on provider configuration attributes.
		ref := pluginConfig.Git.Ref.Value
		if ref == "" {
			ref = "main"
		}

		username, password := p.getGithubCredentials(pluginConfig.Git.Auth)

		gitRepo, err := storagegit.NewSourceCodeRepository(storagegit.SourceCodeRepositoryConfig{
			URL:          pluginConfig.Git.URL.Value,
			BranchOrTag:  ref,
			Dir:          pluginConfig.Git.Dir.Value,
			AuthUsername: username,
			AuthPassword: password,
		})
		if err != nil {
			return nil, fmt.Errorf("could not obtain source code from git repository: %w", err)
		}

		return gitRepo, nil
	}

	// Invalid source code repo.
	return nil, fmt.Errorf("plugin source code source missing")
}

func (p *tfProvider) getGithubCredentials(auth *providerDataPluginV1SourceGitAuth) (username, password string) {
	// Auth disabled.
	if auth == nil {
		return "", ""
	}

	username = os.Getenv("GOPLUGIN_GIT_USERNAME")
	password = os.Getenv("GOPLUGIN_GIT_PASSWORD")

	// If we have explicit config, this has priority over env vars.
	cfgUser := auth.Username.Value
	if cfgUser != "" {
		username = cfgUser
	}
	cfgPassword := auth.Password.Value
	if cfgPassword != "" {
		password = cfgPassword
	}

	// Helper for github access:
	//
	// Github doesn't need the user while using a token, however, an empty user will fail,
	// so if we have password and an empty username, we set a fake user.
	if password != "" && username == "" {
		username = "guybrush-threepwood"
	}

	return username, password
}
