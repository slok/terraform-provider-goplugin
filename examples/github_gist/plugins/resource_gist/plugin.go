package tf

import (
	"context"
	"encoding/json"
	"fmt"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
)

type plugin struct {
	gistManager GistManager
}

func NewResourcePlugin(config string) (apiv1.ResourcePlugin, error) {
	c := Configuration{}
	err := json.Unmarshal([]byte(config), &c)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON configuration: %w", err)
	}

	err = c.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return plugin{
		gistManager: NewAPIGistManager(c.GithubToken, c.APIURL),
	}, nil
}

func (p plugin) CreateResource(ctx context.Context, r apiv1.CreateResourceRequest) (*apiv1.CreateResourceResponse, error) {
	rd, err := tfResourceDataToModel(r.ResourceData)
	if err != nil {
		return nil, fmt.Errorf("could not load resource data from Terraform: %w", err)
	}

	model := Gist{ResourceData: *rd}
	newModel, err := p.gistManager.CreateGist(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("could not create gist: %w", err)
	}

	return &apiv1.CreateResourceResponse{
		ID: newModel.ID,
	}, nil
}

func (p plugin) ReadResource(ctx context.Context, r apiv1.ReadResourceRequest) (*apiv1.ReadResourceResponse, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("id missing, is required")
	}

	model, err := p.gistManager.GetGist(ctx, r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get gist: %w", err)
	}

	res, err := json.Marshal(model.ResourceData)
	if err != nil {
		return nil, fmt.Errorf("could not marshal into JSON: %w", err)
	}

	return &apiv1.ReadResourceResponse{
		ResourceData: string(res),
	}, nil
}

func (p plugin) DeleteResource(ctx context.Context, r apiv1.DeleteResourceRequest) (*apiv1.DeleteResourceResponse, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("id missing, is required")
	}

	err := p.gistManager.DeleteGist(ctx, r.ID)
	if err != nil {
		return nil, fmt.Errorf("could not delete gist: %w", err)
	}

	return &apiv1.DeleteResourceResponse{}, nil
}

func (p plugin) UpdateResource(ctx context.Context, r apiv1.UpdateResourceRequest) (*apiv1.UpdateResourceResponse, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("id missing, is required")
	}

	rdPlan, err := tfResourceDataToModel(r.ResourceData)
	if err != nil {
		return nil, fmt.Errorf("could not load resource data from Terraform Plan: %w", err)
	}

	rdState, err := tfResourceDataToModel(r.ResourceDataState)
	if err != nil {
		return nil, fmt.Errorf("could not load resource data from Terraform State: %w", err)
	}

	// github doesn't allow changing visibility.
	if rdState.Public != rdPlan.Public {
		return nil, fmt.Errorf("visibility of gist can't be changed once created")
	}

	// To be able to delete files we need to set the file as empty content.
	for name := range rdState.Files {
		// If not present in plan, and is in state, means we need to add it as empty so the manager deletes it.
		_, ok := rdPlan.Files[name]
		if !ok {
			rdPlan.Files[name] = ""
		}
	}

	model := Gist{
		ID:           r.ID,
		ResourceData: *rdPlan,
	}
	err = p.gistManager.UpdateGist(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("could not update gist: %w", err)
	}

	return &apiv1.UpdateResourceResponse{}, nil
}
