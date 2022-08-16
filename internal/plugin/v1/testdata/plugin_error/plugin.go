package tf

import (
	"context"
	"fmt"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewResourcePlugin(opts string) (apiv1.ResourcePlugin, error) {
	return plugin{errorMessage: "something"}, nil
}

type plugin struct {
	errorMessage string
}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	return nil, fmt.Errorf(p.errorMessage)
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	return nil, fmt.Errorf(p.errorMessage)
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
	return nil, fmt.Errorf(p.errorMessage)
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
	return nil, fmt.Errorf(p.errorMessage)
}
