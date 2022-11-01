package v1

import "context"

// ReadDataSourceRequest is the request that the plugin will receive on resource `Read` operation.
type ReadDataSourceRequest struct {
	// Attributes is the data the Terraform user will provide to the plugin to manage the resource.
	Attributes string
}

// ReadDataSourceResponse is the response that the plugin will return after resource `Read` operation.
type ReadDataSourceResponse struct {
	// Result is the result the plugin will return to Terraform.
	Result string
}

// DataSourcePlugin knows how to handle a Terraform data source by implementing gathering data operations.
type DataSourcePlugin interface {
	// ReadDataSource will be responsible of:
	//
	// - Using the provided arguments to return a result.
	ReadDataSource(ctx context.Context, r ReadDataSourceRequest) (*ReadDataSourceResponse, error)
}

// DefaultDataSourcePluginFactoryName is the default name used by the plugin engine to search for the plugin factory
// on the plugin source code.
const DefaultDataSourcePluginFactoryName = "NewDataSourcePlugin"

// ResourcePluginFactory is the function type that the plugin engine will load and run to get the plugin that
// will be executed afterwards. E.g:
//
//	func NewDataSourcePlugin(options string) (apiv1.DataSourcePlugin, error) {
//		//...
//		return myPlugin{}, nil
//	}
type DataSourcePluginFactory = func(options string) (DataSourcePlugin, error)
