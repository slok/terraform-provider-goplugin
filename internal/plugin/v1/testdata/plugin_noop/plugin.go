package tf

import (
	"context"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewResourcePlugin(opts string) (apiv1.ResourcePlugin, error) {
	return plugin{}, nil
}

type plugin struct{}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	return &apiv1.CreateResourceResponse{}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	return &apiv1.ReadResourceResponse{}, nil
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
	return &apiv1.DeleteResourceResponse{}, nil
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
	return &apiv1.UpdateResourceResponse{}, nil
}
