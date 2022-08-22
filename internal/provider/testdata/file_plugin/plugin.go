package tf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type Attributes struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (a Attributes) validate() error {
	if a.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

func NewResourcePlugin(config string) (apiv1.ResourcePlugin, error) {
	return plugin{}, nil
}

type plugin struct{}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	att := Attributes{}
	err := json.Unmarshal([]byte(r.Attributes), &att)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = att.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid resource data: %w", err)
	}

	err = os.WriteFile(att.Path, []byte(att.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	return &apiv1.CreateResourceResponse{
		ID: att.Path,
	}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	data, err := os.ReadFile(r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	res, err := json.Marshal(Attributes{
		Path:    r.ID,
		Content: string(data),
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshal into JSON: %w", err)
	}

	return &apiv1.ReadResourceResponse{
		Attributes: string(res),
	}, nil
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("invalid path: %q", r.ID)
	}

	err := os.Remove(r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not delete file: %w", err)
	}

	return &apiv1.DeleteResourceResponse{}, nil
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
	att := Attributes{}
	err := json.Unmarshal([]byte(r.Attributes), &att)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = os.WriteFile(att.Path, []byte(att.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	return &apiv1.UpdateResourceResponse{}, nil
}
