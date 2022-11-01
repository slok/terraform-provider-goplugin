package moduledir_test

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/moduledir"
)

func TestSourceCodeRepository(t *testing.T) {
	tests := map[string]struct {
		dir          string
		expDataFiles map[string]string
		expErr       bool
	}{
		"An invalid Go package should fail.": {
			dir:    "testdata/pkg2",
			expErr: true,
		},

		"Loading a valid package should load the package.": {
			dir: "testdata/pkg1",
			expDataFiles: map[string]string{
				"gopath/src/pkg1/go.mod":         "module pkg1\n",
				"gopath/src/pkg1/test1/f1.go":    "package test1\n\nfunc test1() {}\n",
				"gopath/src/pkg1/test2/test2.go": "package test2\n\nfunc test2() {}\n",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			repo, err := moduledir.NewSourceCodeRepository(os.DirFS(test.dir))

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				gotDataFiles := getFSFiles(repo.FS(context.TODO()))
				assert.Equal(test.expDataFiles, gotDataFiles)

			}
		})
	}
}

func getFSFiles(f fs.FS) map[string]string {
	data := map[string]string{}
	err := fs.WalkDir(f, ".", fs.WalkDirFunc(func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		bs, err := fs.ReadFile(f, path)
		if err != nil {
			return err
		}

		data[path] = string(bs)

		return nil
	}))
	if err != nil {
		panic(err)
	}

	return data
}
