package testing

import (
	"context"
	"fmt"
	"os"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
	pluginv1 "github.com/slok/terraform-provider-goplugin/internal/plugin/v1"
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

// NewTestResourcePlugin is a helper util to load a plugin using the engine that
// will use the terraform provider. In the sense of an acceptance/integration test.
//
// This has benefits over loading the plugin directly with Go, by using this method
// you will be sure that what is executed is what the terraform provider will execute,
// so, if you use a not supported feature or the engine has a bug, this will be
// detected on the tests instead by running Terraform directly.
//
// Note: All plugin files must be at one dir depth level, this is by design on the provider.
func NewTestResourcePlugin(ctx context.Context, pluginDir string, options string) (apiv1.ResourcePlugin, error) {
	// Load plugin source from the file system.
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory files: %w", err)
	}

	data := []string{}
	for _, file := range files {
		d, err := os.ReadFile(file.Name())
		if err != nil {
			return nil, fmt.Errorf("could not read %q file: %w", file.Name(), err)
		}
		data = append(data, string(d))
	}

	// Create resource plugin.
	f := pluginv1.NewFactory()
	return f.NewResourcePlugin(ctx, storage.DataSourceCodeRepository(data), options)
}
