package tf

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type ResourceData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    int    `json:"mode"`
}

func (r ResourceData) validate() error {
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}

	if r.Mode <= 0 {
		return fmt.Errorf("invalid mode")
	}

	return nil
}

func NewResourcePlugin(config string) (apiv1.ResourcePlugin, error) {
	return plugin{}, nil
}

type plugin struct{}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	rd := ResourceData{}
	err := json.Unmarshal([]byte(r.ResourceData), &rd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = rd.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid resource data: %w", err)
	}

	err = os.WriteFile(rd.Path, []byte(rd.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	err = p.chown(rd.Path, rd.Mode)
	if err != nil {
		return nil, fmt.Errorf("could not chown file: %w", err)
	}

	return &apiv1.CreateResourceResponse{
		ID: rd.Path,
	}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	data, err := os.ReadFile(r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	info, err := os.Stat(r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not read file info: %w", err)
	}

	// Make sure we expose the file mode as an octal number.
	modeS := strconv.FormatInt(int64(info.Mode()), 8)
	mode, _ := strconv.ParseInt(modeS, 10, 64)

	res, err := json.Marshal(ResourceData{
		Path:    r.ID,
		Content: string(data),
		Mode:    int(mode),
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshal into JSON: %w", err)
	}

	return &apiv1.ReadResourceResponse{
		ResourceData: string(res),
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
	rd := ResourceData{}
	err := json.Unmarshal([]byte(r.ResourceData), &rd)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = os.WriteFile(rd.Path, []byte(rd.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	err = p.chown(rd.Path, rd.Mode)
	if err != nil {
		return nil, fmt.Errorf("could not chown file: %w", err)
	}

	return &apiv1.UpdateResourceResponse{}, nil
}

func (p plugin) chown(name string, fileMode int) error {
	// Make sure we are working in octal mode, JSON will send us the number as decimal
	// however its an octal.
	modeS := strconv.FormatInt(int64(fileMode), 10)
	modeI, _ := strconv.ParseInt(modeS, 8, 64)
	mode := fs.FileMode(modeI)

	err := os.Chmod(name, mode)
	if err != nil {
		return fmt.Errorf("could not change file mode: %w", err)
	}

	return nil
}
