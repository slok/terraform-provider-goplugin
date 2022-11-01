package murmur3

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spaolacci/murmur3"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewDataSourcePlugin(config string) (apiv1.DataSourcePlugin, error) {
	return dataSourceplugin{}, nil
}

type dataSourceplugin struct{}

type tfAttributes struct {
	Value string `json:"value"`
}

func (t tfAttributes) validate() error {
	if t.Value == "" {
		return fmt.Errorf("value is required")
	}

	return nil
}

type tfResult struct {
	Hash string `json:"hash"`
}

func (d dataSourceplugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
	// Load input data.
	attrs := tfAttributes{}
	err := json.Unmarshal([]byte(r.Attributes), &attrs)
	if err != nil {
		return nil, err
	}

	err = attrs.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid attributes: %w", err)
	}

	// Hash.
	h := murmur3.Sum32([]byte(attrs.Value))
	hashValue := strconv.Itoa(int(h))

	// Output result.
	tRes := tfResult{Hash: hashValue}
	res, err := json.Marshal(tRes)
	if err != nil {
		return nil, err
	}

	return &apiv1.ReadDataSourceResponse{Result: string(res)}, nil
}
