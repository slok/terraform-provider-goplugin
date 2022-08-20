package tf

import (
	"context"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type plugin struct{}

func NewDataSourcePlugin(config string) (apiv1.DataSourcePlugin, error) {
	return plugin{}, nil
}

func (p plugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
	return &apiv1.ReadDataSourceResponse{
		Result: r.Arguments,
	}, nil
}
