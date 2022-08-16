package v1

import "context"

type CreateResourceRequest struct {
	// ResourceData is the data the user must provider to handle correctly.
	ResourceData string
}

type CreateResourceResponse struct {
	// The ID that tracks this resource.
	ID string
}

type ReadResourceRequest struct {
	// The ID that tracks this resource.
	ID string
}

type ReadResourceResponse struct {
	// ResourceData is the data the user must provider to handle correctly.
	ResourceData string
}

type UpdateResourceRequest struct {
	// ResourceData is the data the user must provider to handle correctly.
	ID string
	// ResourceData is the data the user must provider to handle correctly.
	ResourceData string
	// ResourceDataState is the same as ResourceData but is the resource data used on the previous
	// terraform apply execution.
	//
	// This field is not normally used except we want to make decisions based on previous terraform apply
	// changes. (E.g: A resource data attribute can't be changed after the creation, so if changes we return an error)
	ResourceDataState string
}

type UpdateResourceResponse struct{}

type DeleteResourceRequest struct {
	// ResourceData is the data the user must provider to handle correctly.
	ID string
}

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
