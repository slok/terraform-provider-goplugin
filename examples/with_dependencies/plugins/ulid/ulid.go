package ulid

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

func NewDataSourcePlugin(config string) (apiv1.DataSourcePlugin, error) {
	return dataSourceplugin{}, nil
}

type dataSourceplugin struct{}

type tfResult struct {
	ULID string `json:"ulid"`
}

func (d dataSourceplugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
	// Hash.
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	ulid, err := ulid.New(ms, entropy)
	if err != nil {
		return nil, fmt.Errorf("could not create new ULID: %w", err)
	}

	// Output result.
	tRes := tfResult{ULID: ulid.String()}
	res, err := json.Marshal(tRes)
	if err != nil {
		return nil, err
	}

	return &apiv1.ReadDataSourceResponse{Result: string(res)}, nil
}
