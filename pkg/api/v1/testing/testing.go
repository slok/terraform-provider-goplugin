package testing

import (
	"context"
	"fmt"
	"os"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/moduledir"
	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type TestResourcePluginConfig struct {
	PluginDir           string
	PluginFactoryName   string
	PluginConfiguration string
}

func (c *TestResourcePluginConfig) defaults() error {
	if c.PluginDir == "" {
		c.PluginDir = "./"
	}

	if c.PluginFactoryName == "" {
		c.PluginFactoryName = apiv1.DefaultResourcePluginFactoryName
	}

	if c.PluginConfiguration == "" {
		c.PluginConfiguration = "{}"
	}

	return nil
}

// NewTestResourcePlugin is a helper util to load a plugin using the engine that
// will use the terraform provider. In the sense of an acceptance/integration test.
//
// This has benefits over loading the plugin directly with Go, by using this method
// you will be sure that what is executed is what the terraform provider will execute,
// so, if you use a not supported feature or the engine has a bug, this will be
// detected on the tests instead by running Terraform directly.
//
// Note: All plugin files must be at one dir depth level, this is by design on the provider.
func NewTestResourcePlugin(ctx context.Context, config TestResourcePluginConfig) (apiv1.ResourcePlugin, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	repo, err := moduledir.NewSourceCodeRepository(os.DirFS(config.PluginDir))
	if err != nil {
		return nil, fmt.Errorf("could not create source code repo: %w", err)
	}

	return pluginv1.NewEngine().NewResourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    config.PluginFactoryName,
		PluginOptions:        config.PluginConfiguration,
	})
}

type TestDataSourcePluginConfig struct {
	PluginDir           string
	PluginFactoryName   string
	PluginConfiguration string
}

func (c *TestDataSourcePluginConfig) defaults() error {
	if c.PluginDir == "" {
		c.PluginDir = "./"
	}

	if c.PluginFactoryName == "" {
		c.PluginFactoryName = apiv1.DefaultDataSourcePluginFactoryName
	}

	if c.PluginConfiguration == "" {
		c.PluginConfiguration = "{}"
	}

	return nil
}

// NewTestDataSourcePlugin is a helper util to load a plugin using the engine that
// will use the terraform provider. In the sense of an acceptance/integration test.
//
// This has benefits over loading the plugin directly with Go, by using this method
// you will be sure that what is executed is what the terraform provider will execute,
// so, if you use a not supported feature or the engine has a bug, this will be
// detected on the tests instead by running Terraform directly.
//
// Note: All plugin files must be at one dir depth level, this is by design on the provider.
func NewTestDataSourcePlugin(ctx context.Context, config TestDataSourcePluginConfig) (apiv1.DataSourcePlugin, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	repo, err := moduledir.NewSourceCodeRepository(os.DirFS(config.PluginDir))
	if err != nil {
		return nil, fmt.Errorf("could not create source code repo: %w", err)
	}

	return pluginv1.NewEngine().NewDataSourcePlugin(ctx, pluginv1.PluginConfig{
		SourceCodeRepository: repo,
		PluginFactoryName:    config.PluginFactoryName,
		PluginOptions:        config.PluginConfiguration,
	})
}
