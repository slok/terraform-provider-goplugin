package moduledir

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"regexp"
	"testing/fstest"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
)

type repo struct {
	fs    fs.FS
	pkg   string
	index string
}

// NewSourceCodeRepository returns a SourceCodeRepository that will load the
// root directory of a fs in a file system that is ready to be used by yaegi as
// a go module loaded in the gopath.
func NewSourceCodeRepository(dirFS fs.FS) (storage.SourceCodeRepository, error) {
	r := repo{}

	// TODO(slok): Cache.
	err := r.init(dirFS)
	if err != nil {
		return nil, fmt.Errorf("could not initialize repo: %w", err)
	}

	return r, nil
}

func (r repo) FS(ctx context.Context) fs.FS {
	return r.fs
}

func (r repo) Index(ctx context.Context) string {
	return r.index
}

func (r repo) Gopath(ctx context.Context) string {
	return goPath
}

func (r repo) ImportPath(ctx context.Context) string {
	return r.pkg
}

const (
	goPath = "gopath"
)

func (r *repo) init(dirFS fs.FS) (err error) {
	gomod, err := fs.ReadFile(dirFS, "go.mod")
	if err != nil {
		return fmt.Errorf("could not read go.mod: %w", err)
	}

	r.pkg, err = extractModule(string(gomod))
	if err != nil {
		return fmt.Errorf("could not extract module: %w", err)
	}

	srcRoot := fmt.Sprintf("%s/src/%s", goPath, r.pkg)

	mapFS := map[string]*fstest.MapFile{}
	tmpData := ""
	err = fs.WalkDir(dirFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Ignore unwanted files.
		for _, rx := range filesToIgnore {
			if rx.MatchString(path) {
				return nil
			}
		}

		fileData, err := fs.ReadFile(dirFS, path)
		if err != nil {
			return fmt.Errorf("could not read  %s: %w", path, err)
		}

		fileName := fmt.Sprintf("%s/%s", srcRoot, path)

		// Append data file.
		mapFS[fileName] = &fstest.MapFile{Data: fileData}
		tmpData += string(fileData)

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not walk directory: %w", err)
	}

	r.fs = fstest.MapFS(mapFS)
	r.index = fmt.Sprintf("%x", sha256.Sum256([]byte(tmpData)))

	return nil
}

var filesToIgnore = []*regexp.Regexp{
	regexp.MustCompile("vendor/github.com/slok/terraform-provider-goplugin/.*$"),
	regexp.MustCompile("vendor/github.com/traefik/yaegi/.*$"),
	regexp.MustCompile("^.git/.*$"),
}

var moduleRegexp = regexp.MustCompile(`(?m)^module +([^\s]+) *$`)

func extractModule(gomod string) (string, error) {
	// Discover module name.
	moduleMatch := moduleRegexp.FindStringSubmatch(gomod)
	if len(moduleMatch) != 2 {
		return "", fmt.Errorf("invalid go module, could not get module name")
	}

	return moduleMatch[1], nil
}
