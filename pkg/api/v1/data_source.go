package v1

import "context"

type ReadDataSourceRequest struct {
	// Arguments is the data the user must provide so the data source can return a result.
	Arguments string
}

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

const DefaultDataSourcePluginFactoryName = "NewDataSourcePlugin"

// ResourcePluginFactory is the function type that the plugin engine will load and run to get the plugin that
// will be executed afterwards. E.g:
//
//	func NewDataSourcePlugin(options string) (apiv1.DataSourcePlugin, error) {
//		//...
//		return myPlugin{}, nil
//	}
type DataSourcePluginFactory = func(options string) (DataSourcePlugin, error)
