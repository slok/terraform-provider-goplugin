package tf

import (
	"fmt"
	"os"
)

type Configuration struct {
	GithubToken string `json:"github_token"`
	APIURL      string `json:"api_url"`
}

func (c *Configuration) validate() error {
	if c.GithubToken == "" {
		c.GithubToken = os.Getenv("TF_GITHUB_TOKEN")
	}

	if c.GithubToken == "" {
		return fmt.Errorf("github token is required")
	}

	if c.APIURL == "" {
		c.APIURL = "https://api.github.com"
	}

	return nil
}

type Attributes struct {
	Description string            `json:"description"`
	Public      bool              `json:"public"`
	Files       map[string]string `json:"files"`
}

type Gist struct {
	ID string
	Attributes
}

func (r Attributes) validate() error {
	for name, content := range r.Files {
		if name == "" {
			return fmt.Errorf("name of file is required")
		}

		if content == "" {
			return fmt.Errorf("content of file %q is required", name)
		}
	}

	return nil
}
