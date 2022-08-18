package tf_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/slok/terraform-provider-goplugin/pkg/api/v1"
	apiv1testing "github.com/slok/terraform-provider-goplugin/pkg/api/v1/testing"
)

type testMockHandler struct {
	t          *testing.T
	expPath    string
	expMethod  string
	expBody    string
	returnCode int
	returnBody string
}

func (t testMockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(t.t, t.expPath, r.URL.Path)
	assert.Equal(t.t, t.expMethod, r.Method)
	data, err := io.ReadAll(r.Body)
	require.NoError(t.t, err)
	assert.Equal(t.t, t.expBody, string(data))

	w.WriteHeader(t.returnCode)
	_, err = w.Write([]byte(t.returnBody))
	require.NoError(t.t, err)
}

func TestCreateResource(t *testing.T) {
	tests := map[string]struct {
		mock        func(t *testing.T) http.Handler
		request     apiv1.CreateResourceRequest
		expResponse *apiv1.CreateResourceResponse
		expErr      bool
	}{
		"Without valid resource data we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.CreateResourceRequest{
				ResourceData: "{dadsad[sdasdsa",
			},
			expErr: true,
		},

		"If status code is not correct it should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists",
					expMethod:  http.MethodPost,
					expBody:    `{"description":"test-desc","public":true,"files":{"f1":{"content":"d1"},"f2":{"content":"d2"}}}`,
					returnCode: http.StatusTeapot,
					returnBody: "{}",
				}
			},
			request: apiv1.CreateResourceRequest{
				ResourceData: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expErr: true,
		},

		"If server returns an invalid JSON, should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists",
					expMethod:  http.MethodPost,
					expBody:    `{"description":"test-desc","public":true,"files":{"f1":{"content":"d1"},"f2":{"content":"d2"}}}`,
					returnCode: http.StatusCreated,
					returnBody: "[{?sdad·!1",
				}
			},
			request: apiv1.CreateResourceRequest{
				ResourceData: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expErr: true,
		},

		"If Server doesn't return an ID, it should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists",
					expMethod:  http.MethodPost,
					expBody:    `{"description":"test-desc","public":true,"files":{"f1":{"content":"d1"},"f2":{"content":"d2"}}}`,
					returnCode: http.StatusCreated,
					returnBody: `{"id": ""}`,
				}
			},
			request: apiv1.CreateResourceRequest{
				ResourceData: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expErr: true,
		},

		"If server returns an ID, it return a correct result.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists",
					expMethod:  http.MethodPost,
					expBody:    `{"description":"test-desc","public":true,"files":{"f1":{"content":"d1"},"f2":{"content":"d2"}}}`,
					returnCode: http.StatusCreated,
					returnBody: `{"id": "1234567890"}`,
				}
			},
			request: apiv1.CreateResourceRequest{
				ResourceData: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expResponse: &apiv1.CreateResourceResponse{
				ID: "1234567890",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			srv := httptest.NewServer(test.mock(t))
			defer srv.Close()

			config, _ := json.Marshal(map[string]any{
				"github_token": "test",
				"api_url":      srv.URL,
			})

			p, err := apiv1testing.NewTestResourcePlugin(context.TODO(), "./", string(config))
			//p, err := plugin.NewResourcePlugin(string(config)) // Used while tests development the tests to check easily the coverage.
			require.NoError(err)

			gotResp, err := p.CreateResource(context.TODO(), test.request)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, gotResp)
			}
		})
	}
}

func TestReadResource(t *testing.T) {
	tests := map[string]struct {
		mock        func(t *testing.T) http.Handler
		request     apiv1.ReadResourceRequest
		expResponse *apiv1.ReadResourceResponse
		expErr      bool
	}{
		"Without id we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.ReadResourceRequest{
				ID: "",
			},
			expErr: true,
		},

		"If status code is not correct it should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodGet,
					expBody:    ``,
					returnCode: http.StatusTeapot,
					returnBody: `{}`,
				}
			},
			request: apiv1.ReadResourceRequest{
				ID: "1234567890",
			},
			expErr: true,
		},

		"If server returns an invalid JSON, should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodGet,
					expBody:    ``,
					returnCode: http.StatusOK,
					returnBody: `{sdsadsadsadsa[sdsad`,
				}
			},
			request: apiv1.ReadResourceRequest{
				ID: "1234567890",
			},
			expErr: true,
		},

		"A correct response should not fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodGet,
					expBody:    ``,
					returnCode: http.StatusOK,
					returnBody: `{
			 		  "id": "123456789",
			 		  "description": "Hello World",
			 		  "files": {
			 		  	"f1.txt": {"filename": "f1.txt", "content": "This is f1"},
			 		  	"f2.txt": {"filename": "f2.txt", "content": "This is f2"}
			 		  }
			 		}`,
				}
			},
			request: apiv1.ReadResourceRequest{
				ID: "1234567890",
			},
			expResponse: &apiv1.ReadResourceResponse{
				ResourceData: `{"description":"Hello World","public":false,"files":{"f1.txt":"This is f1","f2.txt":"This is f2"}}`,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			srv := httptest.NewServer(test.mock(t))
			defer srv.Close()

			config, _ := json.Marshal(map[string]any{
				"github_token": "test",
				"api_url":      srv.URL,
			})

			p, err := apiv1testing.NewTestResourcePlugin(context.TODO(), "./", string(config))
			//p, err := plugin.NewResourcePlugin(string(config)) // Used while tests development the tests to check easily the coverage.
			require.NoError(err)

			gotResp, err := p.ReadResource(context.TODO(), test.request)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, gotResp)
			}
		})
	}
}

func TestDeleteResource(t *testing.T) {
	tests := map[string]struct {
		mock        func(t *testing.T) http.Handler
		request     apiv1.DeleteResourceRequest
		expResponse *apiv1.DeleteResourceResponse
		expErr      bool
	}{
		"Without id we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.DeleteResourceRequest{
				ID: "",
			},
			expErr: true,
		},

		"If status code is not correct it should fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodDelete,
					expBody:    ``,
					returnCode: http.StatusTeapot,
					returnBody: `{}`,
				}
			},
			request: apiv1.DeleteResourceRequest{
				ID: "1234567890",
			},
			expErr: true,
		},

		"A correct deletion should not fail.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodDelete,
					expBody:    ``,
					returnCode: http.StatusNoContent,
					returnBody: ``,
				}
			},
			request: apiv1.DeleteResourceRequest{
				ID: "1234567890",
			},
			expResponse: &apiv1.DeleteResourceResponse{},
			expErr:      false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			srv := httptest.NewServer(test.mock(t))
			defer srv.Close()

			config, _ := json.Marshal(map[string]any{
				"github_token": "test",
				"api_url":      srv.URL,
			})

			p, err := apiv1testing.NewTestResourcePlugin(context.TODO(), "./", string(config))
			//p, err := plugin.NewResourcePlugin(string(config)) // Used while tests development the tests to check easily the coverage.
			require.NoError(err)

			gotResp, err := p.DeleteResource(context.TODO(), test.request)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, gotResp)
			}
		})
	}
}

func TestUpdateResource(t *testing.T) {
	tests := map[string]struct {
		mock        func(t *testing.T) http.Handler
		request     apiv1.UpdateResourceRequest
		expResponse *apiv1.UpdateResourceResponse
		expErr      bool
	}{
		"Without id we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.UpdateResourceRequest{
				ID:                "",
				ResourceData:      `{}`,
				ResourceDataState: `{}`,
			},
			expErr: true,
		},

		"Without valid resource data we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.UpdateResourceRequest{
				ID:                "1234567890",
				ResourceData:      `{`,
				ResourceDataState: `{}`,
			},
			expErr: true,
		},

		"Without valid state resource data we should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.UpdateResourceRequest{
				ID:                "1234567890",
				ResourceData:      `{}`,
				ResourceDataState: `{`,
			},
			expErr: true,
		},

		"Changing a gist visibility should fail.": {
			mock: func(t *testing.T) http.Handler { return nil },
			request: apiv1.UpdateResourceRequest{
				ID:                "1234567890",
				ResourceData:      `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
				ResourceDataState: `{"description":"test-desc","public":false,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expErr: true,
		},

		"A correct content update should update the content.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodPatch,
					expBody:    `{"description":"test-desc","files":{"f1":{"content":"d1-mutated"},"f2":{"content":"d3"}}}`,
					returnCode: http.StatusOK,
					returnBody: `{}`,
				}
			},
			request: apiv1.UpdateResourceRequest{
				ID:                "1234567890",
				ResourceData:      `{"description":"test-desc","public":true,"files":{"f1": "d1-mutated", "f2": "d3"}}`,
				ResourceDataState: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expResponse: &apiv1.UpdateResourceResponse{},
		},

		"Deleting files from gist should be able to delete files correctly.": {
			mock: func(t *testing.T) http.Handler {
				return testMockHandler{
					t:          t,
					expPath:    "/gists/1234567890",
					expMethod:  http.MethodPatch,
					expBody:    `{"description":"test-desc","files":{"f1":{"content":"d1"},"f2":{"content":""}}}`,
					returnCode: http.StatusOK,
					returnBody: `{}`,
				}
			},
			request: apiv1.UpdateResourceRequest{
				ID:                "1234567890",
				ResourceData:      `{"description":"test-desc","public":true,"files":{"f1": "d1"}}`,
				ResourceDataState: `{"description":"test-desc","public":true,"files":{"f1": "d1", "f2": "d2"}}`,
			},
			expResponse: &apiv1.UpdateResourceResponse{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			srv := httptest.NewServer(test.mock(t))
			defer srv.Close()

			config, _ := json.Marshal(map[string]any{
				"github_token": "test",
				"api_url":      srv.URL,
			})

			p, err := apiv1testing.NewTestResourcePlugin(context.TODO(), "./", string(config))
			//p, err := plugin.NewResourcePlugin(string(config)) // Used while tests development the tests to check easily the coverage.
			require.NoError(err)

			gotResp, err := p.UpdateResource(context.TODO(), test.request)
			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResponse, gotResp)
			}
		})
	}
}
