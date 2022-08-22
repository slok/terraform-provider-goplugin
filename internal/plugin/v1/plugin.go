package v1

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing/fstest"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

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

	pluginSource, err := config.SourceCodeRepository.GetSourceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin source code: %w", err)
	}

	// Sanitize plugin source files like validating and ignoring files that should not be loaded.
	sanitizedPluginSource, err := sanitizedPluginSource(ctx, pluginSource)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin source: %w", err)
	}

	// Get plugin from cache if we already have it.
	index := pluginIndex(ctx, sanitizedPluginSource, config.PluginOptions, config.PluginFactoryName)
	p, ok := e.resourcePluginsCache.Load(index)
	if ok {
		// Should always be a resource plugin, we control the type internally,
		// panicking its ok, shouldn't happen.
		plugin := p.(apiv1.ResourcePlugin)
		return plugin, nil
	}

	// Create Yaegi plugin.
	pluginFactory, err := loadRawResourcePluginFactory(ctx, config.PluginFactoryName, sanitizedPluginSource)
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

	pluginSource, err := config.SourceCodeRepository.GetSourceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin source code: %w", err)
	}

	// Sanitize plugin source files like validating and ignoring files that should not be loaded.
	sanitizedPluginSource, err := sanitizedPluginSource(ctx, pluginSource)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin source: %w", err)
	}

	// Get plugin from cache if we already have it.
	index := pluginIndex(ctx, sanitizedPluginSource, config.PluginOptions, config.PluginFactoryName)
	p, ok := e.dataSourcePluginsCache.Load(index)
	if ok {
		// Should always be a data source plugin, we control the type internally,
		// panicking its ok, shouldn't happen.
		plugin := p.(apiv1.DataSourcePlugin)
		return plugin, nil
	}

	// Create Yaegi plugin.
	pluginFactory, err := loadRawDataSourcePluginFactory(ctx, config.PluginFactoryName, sanitizedPluginSource)
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

var packageRegexp = regexp.MustCompile(`(?m)^package +([^\s]+) *$`)

func sanitizedPluginSource(ctx context.Context, pluginSource []string) ([]string, error) {
	newSrc := []string{}
	for _, src := range pluginSource {
		// Discover package name.
		packageMatch := packageRegexp.FindStringSubmatch(src)
		if len(packageMatch) != 2 {
			return nil, fmt.Errorf("invalid plugin source code, could not get package name")
		}
		pkg := packageMatch[1]

		// Ignore test package files so we simplify plugin loading with two different
		//packages (and don't make yaegi fail).
		if strings.HasSuffix(pkg, "_test") {
			continue
		}

		newSrc = append(newSrc, src)
	}

	return newSrc, nil
}

func pluginIndex(ctx context.Context, pluginSource []string, pluginOptions string, factoryName string) string {
	// Wrap all the plugin information as a single string bundle.
	sort.Strings(pluginSource)
	bundle := strings.Join(pluginSource, "") + pluginOptions + factoryName

	// Get plugin bundle SHA.
	sha := sha256.Sum256([]byte(bundle))

	return fmt.Sprintf("%x", sha)
}

const pluginMemFSDir = "plugin"

func loadRawResourcePluginFactory(ctx context.Context, pluginFactoryName string, srcs []string) (apiv1.ResourcePluginFactory, error) {
	yaegiInterp, err := newPluginYaegiInterpreter(ctx, srcs, pluginMemFSDir)
	if err != nil {
		return nil, fmt.Errorf("could not create Yaegi interpreter: %w", err)
	}

	// Get plugin logic by
	// - Importing the memfs local dir package.
	// - Getting the plugin factory name convention.
	// - Type assert to our factory type convention.
	pluginImportStatement := fmt.Sprintf(`import plugin "./%s"`, pluginMemFSDir)
	_, err = yaegiInterp.EvalWithContext(ctx, pluginImportStatement)
	if err != nil {
		return nil, fmt.Errorf("could not import plugin package: %w", err)
	}

	srcPluginIdentifier := fmt.Sprintf("plugin.%s", pluginFactoryName)
	pluginFuncTmp, err := yaegiInterp.EvalWithContext(ctx, srcPluginIdentifier)
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	pluginFunc, ok := pluginFuncTmp.Interface().(apiv1.ResourcePluginFactory)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	return pluginFunc, nil
}

func loadRawDataSourcePluginFactory(ctx context.Context, pluginFactoryName string, srcs []string) (apiv1.DataSourcePluginFactory, error) {
	yaegiInterp, err := newPluginYaegiInterpreter(ctx, srcs, pluginMemFSDir)
	if err != nil {
		return nil, fmt.Errorf("could not create Yaegi interpreter: %w", err)
	}

	// Get plugin logic by
	// - Importing the memfs local dir package.
	// - Getting the plugin factory name convention.
	// - Type assert to our factory type convention.
	pluginImportStatement := fmt.Sprintf(`import plugin "./%s"`, pluginMemFSDir)
	_, err = yaegiInterp.EvalWithContext(ctx, pluginImportStatement)
	if err != nil {
		return nil, fmt.Errorf("could not import plugin package: %w", err)
	}

	srcPluginIdentifier := fmt.Sprintf("plugin.%s", pluginFactoryName)
	pluginFuncTmp, err := yaegiInterp.EvalWithContext(ctx, srcPluginIdentifier)
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
func newPluginYaegiInterpreter(ctx context.Context, srcs []string, pluginDir string) (*interp.Interpreter, error) {
	if pluginDir == "" {
		return nil, fmt.Errorf("plugin directory to set the go plugin on the FS is required")
	}

	// Create our plugin memory FS with the plugin sources in a plugin dir.
	mapFS := map[string]*fstest.MapFile{}
	for i, src := range srcs {
		fileName := fmt.Sprintf("%s/%d.go", pluginDir, i)

		mapFS[fileName] = &fstest.MapFile{Data: []byte(src)}
	}

	// Create interpreter
	i := interp.New(interp.Options{
		SourcecodeFilesystem: fstest.MapFS(mapFS),
		Env:                  os.Environ(),
	})

	// Add standard library.
	err := i.Use(stdlib.Symbols)
	if err != nil {
		return nil, fmt.Errorf("yaegi could not use stdlib symbols: %w", err)
	}

	// Add our own plugin library.
	err = i.Use(yaegicustom.Symbols)
	if err != nil {
		return nil, fmt.Errorf("yaegi could not use custom symbols: %w", err)
	}

	return i, nil
}
