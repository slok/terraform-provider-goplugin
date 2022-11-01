package git_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/git"
)

func TestSourceCodeRepository(t *testing.T) {
	tests := map[string]struct {
		config git.SourceCodeRepositoryConfig
		expErr bool
	}{
		"Missing repository should fail.": {
			config: git.SourceCodeRepositoryConfig{
				URL: "https://github.com/slok/something-missing.git",
			},
			expErr: true,
		},

		"Cloning a valid repo not being valid go package, should fail.": {
			config: git.SourceCodeRepositoryConfig{
				URL:         "https://github.com/slok/custom-css",
				BranchOrTag: "main",
			},
			expErr: true,
		},

		"Cloning a repo by matching multiple files should return the data.": {
			config: git.SourceCodeRepositoryConfig{
				URL:         "https://github.com/oklog/run",
				BranchOrTag: "v1.1.0",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo, err := git.NewSourceCodeRepository(test.config)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.NotNil(repo.FS(context.TODO()))
			}
		})
	}
}
