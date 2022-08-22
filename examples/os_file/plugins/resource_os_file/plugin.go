package tf

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type Attributes struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Mode    int    `json:"mode"`
}

func (a Attributes) validate() error {
	if a.Path == "" {
		return fmt.Errorf("path is required")
	}

	if a.Mode <= 0 {
		return fmt.Errorf("invalid mode")
	}

	return nil
}

func NewResourcePlugin(config string) (apiv1.ResourcePlugin, error) {
	return plugin{}, nil
}

type plugin struct{}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	atts := Attributes{}
	err := json.Unmarshal([]byte(r.Attributes), &atts)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = atts.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid resource data: %w", err)
	}

	err = os.WriteFile(atts.Path, []byte(atts.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	err = p.chown(atts.Path, atts.Mode)
	if err != nil {
		return nil, fmt.Errorf("could not chown file: %w", err)
	}

	return &apiv1.CreateResourceResponse{
		ID: atts.Path,
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

	res, err := json.Marshal(Attributes{
		Path:    r.ID,
		Content: string(data),
		Mode:    int(mode),
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
	atts := Attributes{}
	err := json.Unmarshal([]byte(r.Attributes), &atts)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = os.WriteFile(atts.Path, []byte(atts.Content), 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write file: %w", err)
	}

	err = p.chown(atts.Path, atts.Mode)
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

type DataSourceArguments struct {
	Path string `json:"path"`
}

func (d DataSourceArguments) validate() error {
	if d.Path == "" {
		return fmt.Errorf("path is required")
	}

	return nil
}

type DataSourceResult struct {
	SizeBytes int64     `json:"size_bytes"`
	Mode      int       `json:"mode"`
	ModTime   time.Time `json:"mod_time"`
	IsDir     bool      `json:"is_dir"`
	Content   string    `json:"content"`
}

func NewDataSourcePlugin(config string) (apiv1.DataSourcePlugin, error) {
	return plugin{}, nil
}

func (p plugin) ReadDataSource(ctx context.Context, r apiv1.ReadDataSourceRequest) (*apiv1.ReadDataSourceResponse, error) {
	args := DataSourceArguments{}
	err := json.Unmarshal([]byte(r.Attributes), &args)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON data: %w", err)
	}

	err = args.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	content, err := os.ReadFile(args.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	info, err := os.Stat(args.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read file info: %w", err)
	}

	// Make sure we expose the file mode as an octal number.
	modeS := strconv.FormatInt(int64(info.Mode()), 8)
	mode, _ := strconv.ParseInt(modeS, 10, 64)

	result, err := json.Marshal(DataSourceResult{
		SizeBytes: info.Size(),
		Mode:      int(mode),
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		Content:   string(content),
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshal into JSON: %w", err)
	}

	return &apiv1.ReadDataSourceResponse{
		Result: string(result),
	}, nil
}
