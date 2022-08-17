package git_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/git"
)

func TestSourceCodeRepository(t *testing.T) {
	tests := map[string]struct {
		config  git.SourceCodeRepositoryConfig
		expData []string
		expErr  bool
	}{
		"Missing repository should fail.": {
			config: git.SourceCodeRepositoryConfig{
				URL: "https://github.com/slok/something-missing.git",
			},
			expErr: true,
		},

		"Cloning a repo by matching multiple files should return the data.": {
			config: git.SourceCodeRepositoryConfig{
				URL:         "https://github.com/slok/simple-ingress-external-auth",
				BranchOrTag: "v0.3.0",
				MatchRegexes: []*regexp.Regexp{
					regexp.MustCompile(`^/scripts/[^/]*\.sh$`),
					regexp.MustCompile(`^/\.github/CODEOWNERS$`),
				},
			},
			expData: []string{
				"*       @slok\n\n",
				"#!/usr/bin/env sh\n\nset -o errexit\nset -o nounset\n\ngo mod tidy",
				"#!/usr/bin/env sh\n\nset -o errexit\nset -o nounset\n\ngo generate ./...",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			repo, err := git.NewSourceCodeRepository(test.config)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				gotData, err := repo.GetSourceCode(context.TODO())
				require.NoError(err)
				assert.Equal(test.expData, gotData)
			}
		})
	}
}
