package tf

import (
	"context"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func (p plugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
	return &apiv1.ReadDataSourceResponse{
		Result: r.Arguments + "from_data_source",
	}, nil
}
