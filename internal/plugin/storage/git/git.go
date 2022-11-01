package git

import (
	"fmt"
	"io/fs"
	"path"
	"testing/fstest"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage/moduledir"
)

type SourceCodeRepositoryConfig struct {
	URL          string
	BranchOrTag  string
	Dir          string
	AuthUsername string
	AuthPassword string
}

func (c *SourceCodeRepositoryConfig) defaults() error {
	if c.URL == "" {
		return fmt.Errorf("git url is required")
	}

	if c.BranchOrTag == "" {
		return fmt.Errorf("ref is required")
	}

	if c.Dir == "" {
		c.Dir = "/"
	}

	if !path.IsAbs(c.Dir) {
		return fmt.Errorf("repo dir should be absolute from the repo root")
	}

	return nil
}

// NewSourceCodeRepository returns a Git based SourceCodeRepository.
func NewSourceCodeRepository(config SourceCodeRepositoryConfig) (storage.SourceCodeRepository, error) {
	err := config.defaults()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	repoFS, err := getRepositoryOnFilesystem(config)
	if err != nil {
		return nil, fmt.Errorf("could not get repo file system: %w", err)
	}

	mapFS := map[string]*fstest.MapFile{}
	err = util.Walk(repoFS, config.Dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath := path[1:]

		// Append data file.
		data, err := util.ReadFile(repoFS, path)
		if err != nil {
			return fmt.Errorf("could not read git file: %w", err)
		}

		mapFS[relPath] = &fstest.MapFile{Data: data}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk git repository: %w", err)
	}

	return moduledir.NewSourceCodeRepository(fstest.MapFS(mapFS))
}

// pluginFsCache will have all the cloned repos in the reference that was obtained
// be aware that branch refs can change, however we know that plugins are loaded
// per Terraform execution on the provider setup, so its ok to cache anything.
//
// Note: We don't care about concurrent usage of the map.
var pluginFsCache = map[string]billy.Filesystem{}

func getRepositoryOnFilesystem(config SourceCodeRepositoryConfig) (billy.Filesystem, error) {
	// Try first from cache.
	id := fmt.Sprintf("%s-%s", config.URL, config.BranchOrTag)
	fs, ok := pluginFsCache[id]
	if ok {
		return fs, nil
	}

	// We will try to clone in tag and branch order.
	possibleRefs := []plumbing.ReferenceName{
		plumbing.NewTagReferenceName(config.BranchOrTag),
		plumbing.NewBranchReferenceName(config.BranchOrTag),
	}

	var err error
	for _, ref := range possibleRefs {
		// Filesystem abstraction based on memory.
		memfs := memfs.New()
		storer := memory.NewStorage()

		var auth transport.AuthMethod
		if config.AuthPassword != "" {
			auth = &http.BasicAuth{
				Username: config.AuthUsername,
				Password: config.AuthPassword,
			}
		}

		_, err = git.Clone(storer, memfs, &git.CloneOptions{
			URL:           config.URL,
			Depth:         1,
			ReferenceName: ref,
			Auth:          auth,
		})
		if err == nil {
			// Store in cache.
			pluginFsCache[id] = memfs
			return memfs, nil
		}
	}

	return nil, fmt.Errorf("could not clone repository: %w", err)
}
