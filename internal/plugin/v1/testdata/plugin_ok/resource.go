package tf

import (
	"context"
	"fmt"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	return &apiv1.CreateResourceResponse{
		ID: r.Attributes + "_test1",
	}, nil
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
	if r.ID != "test1" {
		return nil, fmt.Errorf("test failed")
	}

	return &apiv1.DeleteResourceResponse{}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	return &apiv1.ReadResourceResponse{
		Attributes: r.ID + "_test1",
	}, nil
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
	if r.ID != "test1" || r.Attributes != "This is" || r.AttributesState != " a test" {
		return nil, fmt.Errorf("test failed")
	}

	return &apiv1.UpdateResourceResponse{}, nil
}
