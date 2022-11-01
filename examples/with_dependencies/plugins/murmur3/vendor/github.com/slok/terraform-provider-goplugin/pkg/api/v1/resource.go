package v1

import (
	"context"
)

// CreateResourceRequest is the request that the plugin will receive on resource `Create` operation.
type CreateResourceRequest struct {
	// Attributes is the data the Terraform user will provide in the configuration (tf files) to the
	// plugin to manage the resource.
	Attributes string
}

// CreateResourceResponse is the response that the plugin will return after resource `Create` operation.
type CreateResourceResponse struct {
	// ID is the ID that tracks this resource.
	ID string
}

// ReadResourceRequest is the request that the plugin will receive on resource `Read` operation.
type ReadResourceRequest struct {
	// ID is the ID that tracks this resource.
	ID string
}

// ReadResourceResponse is the response that the plugin will return after resource `Read` operation.
type ReadResourceResponse struct {
	// Attributes is the data the Terraform user will provide in the configuration (tf files) to the
	// plugin to manage the resource.
	// On read operation normally this is used to refresh the state and check the attributes provided
	// by the user match the ones that need to be read, then Terraform will watch for drifts.
	Attributes string
}

// UpdateResourceRequest is the request that the plugin will receive on resource `Update` operation.
type UpdateResourceRequest struct {
	// ID is the ID that tracks this resource.
	ID string
	// Attributes is the data the Terraform user will provide in the configuration (tf files) to the
	// plugin to manage the resource.
	Attributes string
	// AttributesState is the same as Attributes but instead of being the configuration from the user, are
	// the attributes used on the previous terraform apply execution.
	//
	// This field is meant to be used used when we want to make decisions based on previous terraform apply changes.
	// (E.g: A resource data attribute can't be changed after the creation, so if changes we return an error)
	AttributesState string
}

// UpdateResourceResponse is the response that the plugin will return after resource `Update` operation.
type UpdateResourceResponse struct{}

// DeleteResourceRequest is the request that the plugin will receive on resource `Delete` operation.
type DeleteResourceRequest struct {
	// ID is the ID that tracks this resource.
	ID string
}

// DeleteResourceResponse is the response that the plugin will return after resource `Delete` operation.
type DeleteResourceResponse struct{}

// ResourcePlugin knows how to handle a Terraform resource by implementing the common Terraform CRUD operations.
type ResourcePlugin interface {
	// CreateResource will be responsible of:
	//
	// - Creating the resource.
	// - Returning the correct ID that will be used from now on to identify the resource.
	// - Generating the ID with the required information to identify a resource (e.g aggregation of 2 properties as a single ID).
	CreateResource(ctx context.Context, r CreateResourceRequest) (*CreateResourceResponse, error)

	// ReadResource will be responsible of:
	//
	// - Using the ID for getting the current real data of the resource (Used on plans and imports).
	ReadResource(ctx context.Context, r ReadResourceRequest) (*ReadResourceResponse, error)

	// UpdateResource will be responsible of:
	//
	// - Using the resource data and ID update the resource if required.
	// - Use the latest applied resource data (state data) to get diffs if required to patch/update specific parts.
	UpdateResource(ctx context.Context, r UpdateResourceRequest) (*UpdateResourceResponse, error)

	// DeleteResource will be responsible of:
	//
	// - Deleting the resource using the ID.
	DeleteResource(ctx context.Context, r DeleteResourceRequest) (*DeleteResourceResponse, error)
}

// DefaultResourcePluginFactoryName is the default name used by the plugin engine to search for the plugin factory
// on the plugin source code.
const DefaultResourcePluginFactoryName = "NewResourcePlugin"

// ResourcePluginFactory is the function type that the plugin engine will load and run to get the plugin that
// will be executed afterwards. E.g:
//
//	func NewResourcePlugin(options string) (apiv1.ResourcePlugin, error) {
//		//...
//		return myPlugin{}, nil
//	}
type ResourcePluginFactory = func(options string) (ResourcePlugin, error)
