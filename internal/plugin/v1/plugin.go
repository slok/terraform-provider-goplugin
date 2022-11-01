package v1

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sync"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/unsafe"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
	"github.com/slok/terraform-provider-goplugin/internal/plugin/v1/yaegicustom"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

// Engine is the plugin engine that knows how to load, prepare and return new plugins.
type Engine struct {
	resourcePluginsCache   sync.Map
	dataSourcePluginsCache sync.Map
}

// NewEngine returns a new plugin V1 engine.
func NewEngine() *Engine {
	return &Engine{}
}

// PluginConfig is the configuration that the engine needs to instantiate a new plugin.
type PluginConfig struct {
	// SourceCodeRepository is the repository where the plugin engine will get the source code for the plugin.
	SourceCodeRepository storage.SourceCodeRepository
	// PluginOptions are the options that will be passed to the plugin factory to create a new plugin.
	PluginOptions string
	// PluginFactoryName is the name that the plugin  engine will search for in the plugin factory inside
	// the plugin source code. It must meet the plugin factory signature.
	// E.g: NewResourcePlugin, NewDataSourcePlugin...
	PluginFactoryName string
}

func (p *PluginConfig) defaults() error {
	if p.SourceCodeRepository == nil {
		return fmt.Errorf("source code repository is required")
	}

	if p.PluginFactoryName == "" {
		return fmt.Errorf("The name of the plugin factory is required")
	}

	return nil
}

// NewResourcePlugin returns a new plugin based on the plugin source code and the plugin options that will be passed on plugin creation.
// the resulting plugin will be able to be used.
func (e *Engine) NewResourcePlugin(ctx context.Context, config PluginConfig) (apiv1.ResourcePlugin, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	// Get plugin from cache if we already have it.
	index := pluginIndex(ctx, config.SourceCodeRepository, config.PluginOptions, config.PluginFactoryName)
	p, ok := e.resourcePluginsCache.Load(index)
	if ok {
		// Should always be a resource plugin, we control the type internally,
		// panicking its ok, shouldn't happen.
		plugin := p.(apiv1.ResourcePlugin)
		return plugin, nil
	}

	// Create Yaegi plugin.
	pluginFactory, err := loadRawResourcePluginFactory(ctx, config.SourceCodeRepository, config.PluginFactoryName)
	if err != nil {
		return nil, fmt.Errorf("could not load plugin: %w", err)
	}

	plugin, err := pluginFactory(config.PluginOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create plugin: %w", err)
	}

	// Store plugin in cache.
	e.resourcePluginsCache.Store(index, plugin)

	return plugin, nil
}

func (e *Engine) NewDataSourcePlugin(ctx context.Context, config PluginConfig) (apiv1.DataSourcePlugin, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	// Get plugin from cache if we already have it.
	index := pluginIndex(ctx, config.SourceCodeRepository, config.PluginOptions, config.PluginFactoryName)
	p, ok := e.dataSourcePluginsCache.Load(index)
	if ok {
		// Should always be a data source plugin, we control the type internally,
		// panicking its ok, shouldn't happen.
		plugin := p.(apiv1.DataSourcePlugin)
		return plugin, nil
	}

	// Create Yaegi plugin.
	pluginFactory, err := loadRawDataSourcePluginFactory(ctx, config.SourceCodeRepository, config.PluginFactoryName)
	if err != nil {
		return nil, fmt.Errorf("could not load plugin: %w", err)
	}

	plugin, err := pluginFactory(config.PluginOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create plugin: %w", err)
	}

	// Store plugin in cache.
	e.resourcePluginsCache.Store(index, plugin)

	return plugin, nil
}

func pluginIndex(ctx context.Context, repo storage.SourceCodeRepository, pluginOptions string, factoryName string) string {
	bundle := repo.Index(ctx) + pluginOptions + factoryName
	sha := sha256.Sum256([]byte(bundle))

	return fmt.Sprintf("%x", sha)
}

const pluginMemFSDir = "plugin"

func loadRawResourcePluginFactory(ctx context.Context, repo storage.SourceCodeRepository, pluginFactoryName string) (apiv1.ResourcePluginFactory, error) {
	yaegiInterp, err := newPluginYaegiInterpreter(ctx, repo, pluginMemFSDir)
	if err != nil {
		return nil, fmt.Errorf("could not create Yaegi interpreter: %w", err)
	}

	importStatement := fmt.Sprintf(`import plugin "%s"`, repo.ImportPath(ctx))
	_, err = yaegiInterp.EvalWithContext(ctx, importStatement)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	// Get plugin logic.
	pluginFuncTmp, err := yaegiInterp.EvalWithContext(ctx, "plugin."+pluginFactoryName)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	pluginFunc, ok := pluginFuncTmp.Interface().(apiv1.ResourcePluginFactory)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	return pluginFunc, nil
}

func loadRawDataSourcePluginFactory(ctx context.Context, repo storage.SourceCodeRepository, pluginFactoryName string) (apiv1.DataSourcePluginFactory, error) {
	yaegiInterp, err := newPluginYaegiInterpreter(ctx, repo, pluginMemFSDir)
	if err != nil {
		return nil, fmt.Errorf("could not create Yaegi interpreter: %w", err)
	}

	importStatement := fmt.Sprintf(`import plugin "%s"`, repo.ImportPath(ctx))
	_, err = yaegiInterp.EvalWithContext(ctx, importStatement)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	// Get plugin logic.
	pluginFuncTmp, err := yaegiInterp.EvalWithContext(ctx, "plugin."+pluginFactoryName)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	pluginFunc, ok := pluginFuncTmp.Interface().(apiv1.DataSourcePluginFactory)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	return pluginFunc, nil
}

// newPluginReadyYaegiInterpreter will:
// - Create a new Yaegi interpreter.
// - Setup a memory based FS with the plugin source code loaded in the specified plugin dir.
// - Add the required libraries available (standard library and our own library).
func newPluginYaegiInterpreter(ctx context.Context, repo storage.SourceCodeRepository, pluginDir string) (*interp.Interpreter, error) {
	// Create interpreter
	i := interp.New(interp.Options{
		SourcecodeFilesystem: repo.FS(ctx),
		Env:                  os.Environ(),
		GoPath:               repo.Gopath(ctx),
	})

	// Add standard library.
	err := i.Use(stdlib.Symbols)
	if err != nil {
		return nil, fmt.Errorf("yaegi could not use stdlib symbols: %w", err)
	}

	// Add unsafe library.
	err = i.Use(unsafe.Symbols)
	if err != nil {
		return nil, fmt.Errorf("yaegi could not use stdlib unsafe symbols: %w", err)
	}

	// Add our own plugin library.
	err = i.Use(yaegicustom.Symbols)
	if err != nil {
		return nil, fmt.Errorf("yaegi could not use custom symbols: %w", err)
	}

	return i, nil
}
