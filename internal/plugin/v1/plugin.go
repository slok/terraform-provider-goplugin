package v1

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing/fstest"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/v1/yaegicustom"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type Factory struct {
	pluginsCache sync.Map
}

// NewFactory returns a new plugin V1 factory.
func NewFactory() Factory {
	return Factory{}
}

// NewResourcePlugin returns a new plugin based on the plugin source code and the plugin options that will be passed on plugin creation.
// the resulting plugin will be able to be used.
func (f *Factory) NewResourcePlugin(ctx context.Context, pluginSource []string, pluginOptions string) (apiv1.ResourcePlugin, error) {
	// Sanitize plugin source files like validating and ignoring files that should not be loaded.
	sanitizedPluginSource, err := f.sanitizedPluginSource(ctx, pluginSource)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin source: %w", err)
	}

	// Get plugin from cache if we already have it.
	index := f.pluginIndex(ctx, sanitizedPluginSource, pluginOptions)
	p, ok := f.pluginsCache.Load(index)
	if ok {
		// Should always be a resource plugin, we control the type internally,
		// panicking its ok, shouldn't happen.
		plugin := p.(apiv1.ResourcePlugin)
		return plugin, nil
	}

	// Create Yaegi plugin.
	pluginFactory, err := loadRawResourcePluginFactory(ctx, sanitizedPluginSource)
	if err != nil {
		return nil, fmt.Errorf("could not load plugin: %w", err)
	}

	plugin, err := pluginFactory(pluginOptions)
	if err != nil {
		return nil, fmt.Errorf("could not create plugin: %w", err)
	}

	// Store plugin in cache.
	f.pluginsCache.Store(index, plugin)

	return plugin, nil
}

var packageRegexp = regexp.MustCompile(`(?m)^package +([^\s]+) *$`)

func (f *Factory) sanitizedPluginSource(ctx context.Context, pluginSource []string) ([]string, error) {
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

func (f *Factory) pluginIndex(ctx context.Context, pluginSource []string, pluginOptions string) string {
	// Wrap all the plugin as a single string.
	sort.Strings(pluginSource)
	allPlugin := strings.Join(pluginSource, "\n") + pluginOptions

	// Get plugin SHA.
	sha := sha256.Sum256([]byte(allPlugin))

	return fmt.Sprintf("%x", sha)
}

// newResourcePlugin is the function the plugins need to have to be able to instantiate a
// new plugin.
type newResourcePlugin = func(options string) (apiv1.ResourcePlugin, error)

func loadRawResourcePluginFactory(ctx context.Context, srcs []string) (newResourcePlugin, error) {
	// Create our plugin memory FS with the plugin sources in a plugin dir.
	mapFS := map[string]*fstest.MapFile{}
	for i, src := range srcs {
		fileName := fmt.Sprintf("plugin/%d.go", i)

		mapFS[fileName] = &fstest.MapFile{Data: []byte(src)}
	}

	// Load the plugin in a new interpreter using the memory file system with our plugin dir.
	// For each plugin we need to use an independent interpreter to avoid name collisions.
	yaegiInterp, err := newYaegiInterpreter(fstest.MapFS(mapFS))
	if err != nil {
		return nil, fmt.Errorf("could not create a new Yaegi interpreter: %w", err)
	}

	// Get plugin logic by
	// - Importing the memfs local dir package
	// - Getting the factory name convention (`NewResourcePlugin`).
	// - Type assert to our factory type convention.
	_, err = yaegiInterp.EvalWithContext(ctx, `import plugin "./plugin"`)
	if err != nil {
		return nil, fmt.Errorf("could not import plugin package: %w", err)
	}
	pluginFuncTmp, err := yaegiInterp.EvalWithContext(ctx, "plugin.NewResourcePlugin")
	if err != nil {
		return nil, fmt.Errorf("could not get plugin: %w", err)
	}

	pluginFunc, ok := pluginFuncTmp.Interface().(newResourcePlugin)
	if !ok {
		return nil, fmt.Errorf("invalid plugin type")
	}

	return pluginFunc, nil
}

func newYaegiInterpreter(fs fs.FS) (*interp.Interpreter, error) {
	i := interp.New(interp.Options{
		SourcecodeFilesystem: fs,
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
