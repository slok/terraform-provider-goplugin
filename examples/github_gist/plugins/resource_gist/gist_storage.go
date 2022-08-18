package tf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GistManager knows how to manage Gists with external storage.
type GistManager interface {
	CreateGist(ctx context.Context, gist Gist) (*Gist, error)
	GetGist(ctx context.Context, id string) (*Gist, error)
	DeleteGist(ctx context.Context, id string) error
	UpdateGist(ctx context.Context, gist Gist) error
}

type apiGistManager struct {
	apiURL  string
	ghToken string
	httpCli *http.Client
}

// NewAPIGistManager returns a Gist manager implementation using Github's HTTP
// rest API. More info here: https://docs.github.com/en/rest/gists/gists.
func NewAPIGistManager(ghToken string, apiURL string) GistManager {
	return apiGistManager{
		apiURL:  apiURL,
		ghToken: ghToken,
		httpCli: http.DefaultClient,
	}
}

func (a apiGistManager) CreateGist(ctx context.Context, gist Gist) (*Gist, error) {
	const (
		path   = "/gists"
		method = http.MethodPost
	)

	type reqAPIModelFiles struct {
		Content string `json:"content"`
	}

	type reqAPIModel struct {
		Description string                      `json:"description"`
		Public      bool                        `json:"public"`
		Files       map[string]reqAPIModelFiles `json:"files"`
	}

	// Map plan request.
	fs := map[string]reqAPIModelFiles{}
	for name, content := range gist.Files {
		fs[name] = reqAPIModelFiles{Content: content}
	}

	data, err := json.Marshal(reqAPIModel{
		Description: gist.Description,
		Public:      gist.Public,
		Files:       fs,
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshall to gist API model: %w", err)
	}

	// Make REST API call.
	req, err := http.NewRequestWithContext(ctx, method, a.apiURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("could not create create API request: %w", err)
	}

	code, body, err := a.httpRequestDo(req)
	if err != nil {
		return nil, fmt.Errorf("create on API failed: %w", err)
	}

	if code != http.StatusCreated {
		return nil, fmt.Errorf("expected status code %d, got %d: %s", http.StatusCreated, code, body)
	}

	type respAPIModel struct {
		ID string `json:"id"`
	}

	mResp := respAPIModel{}
	err = json.Unmarshal([]byte(body), &mResp)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON response body: %w", err)
	}

	if mResp.ID == "" {
		return nil, fmt.Errorf("gist id is missing")
	}

	gist.ID = mResp.ID

	return &gist, nil
}

func (a apiGistManager) GetGist(ctx context.Context, id string) (*Gist, error) {
	const (
		path   = "/gists/%s"
		method = http.MethodGet
	)

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	// Make REST API call.
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf(a.apiURL+path, id), nil)
	if err != nil {
		return nil, fmt.Errorf("could not get API request: %w", err)
	}

	code, body, err := a.httpRequestDo(req)
	if err != nil {
		return nil, fmt.Errorf("get on API failed: %w", err)
	}

	if code != http.StatusOK {
		return nil, fmt.Errorf("expected status code %d, got %d: %s", http.StatusOK, code, body)
	}

	type respAPIModelFiles struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}

	type respAPIModel struct {
		ID          string                       `json:"id"`
		Title       string                       `json:"title"`
		Description string                       `json:"description"`
		Public      bool                         `json:"public"`
		Files       map[string]respAPIModelFiles `json:"files"`
	}
	mResp := respAPIModel{}
	err = json.Unmarshal([]byte(body), &mResp)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON response body: %w", err)
	}

	// Map request.
	fs := map[string]string{}
	for _, f := range mResp.Files {
		fs[f.Filename] = f.Content
	}

	gist := Gist{
		ID: mResp.ID,
		ResourceData: ResourceData{
			Description: mResp.Description,
			Public:      mResp.Public,
			Files:       fs,
		},
	}

	return &gist, nil
}

func (a apiGistManager) DeleteGist(ctx context.Context, id string) error {
	const (
		path   = "/gists/%s"
		method = http.MethodDelete
	)

	if id == "" {
		return fmt.Errorf("ID is required")
	}

	// Make REST API call.
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf(a.apiURL+path, id), nil)
	if err != nil {
		return fmt.Errorf("could not get API request: %w", err)
	}

	code, body, err := a.httpRequestDo(req)
	if err != nil {
		return fmt.Errorf("delete on API failed: %w", err)
	}

	if code != http.StatusNoContent {
		return fmt.Errorf("expected status code %d, got %d: %s", http.StatusNoContent, code, body)
	}

	return nil
}

// UpdateGist will update the the gist, if you need to delete files, you will
// need to provide them with the title and the content as empty.
func (a apiGistManager) UpdateGist(ctx context.Context, gist Gist) error {
	const (
		path   = "/gists/%s"
		method = http.MethodPatch
	)

	type reqAPIModelFiles struct {
		Content string `json:"content"`
	}

	type reqAPIModel struct {
		Description string                      `json:"description"`
		Files       map[string]reqAPIModelFiles `json:"files"`
	}

	// Map request.
	fs := map[string]reqAPIModelFiles{}
	for name, content := range gist.Files {
		fs[name] = reqAPIModelFiles{Content: content}
	}

	data, err := json.Marshal(reqAPIModel{
		Description: gist.Description,
		Files:       fs,
	})
	if err != nil {
		return fmt.Errorf("could not marshall to gist API model: %w", err)
	}

	// Make REST API call.
	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf(a.apiURL+path, gist.ID), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("could not update API request: %w", err)
	}

	code, body, err := a.httpRequestDo(req)
	if err != nil {
		return fmt.Errorf("update on API failed: %w", err)
	}

	if code != http.StatusOK {
		return fmt.Errorf("expected status code %d, got %d: %s", http.StatusOK, code, body)
	}

	return nil
}

// We use this instead of creating our own Transport, because of yaegi panic problems.
func (a apiGistManager) httpRequestDo(req *http.Request) (int, string, error) {
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", a.ghToken))

	resp, err := a.httpCli.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("could not read response body: %w", err)
	}

	return resp.StatusCode, string(respBody), nil
}
