package tf

import (
	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewResourcePlugin(configuration string) (apiv1.ResourcePlugin, error) {
	return plugin{}, nil
}

func NewDataSourcePlugin(configuration string) (apiv1.DataSourcePlugin, error) {
	return plugin{}, nil
}

type plugin struct{}
