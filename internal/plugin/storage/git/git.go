package git

import (
	"fmt"
	"io/fs"
	"regexp"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/slok/terraform-provider-goplugin/internal/plugin/storage"
)

type SourceCodeRepositoryConfig struct {
	URL          string
	BranchOrTag  string
	MatchRegexes []*regexp.Regexp
}

func (c *SourceCodeRepositoryConfig) defaults() error {
	if c.URL == "" {
		return fmt.Errorf("git url is required")
	}

	if c.BranchOrTag == "" {
		c.BranchOrTag = "main"
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

	files := []string{}
	err = util.Walk(repoFS, "/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if directory or doesn't match with expected file.
		if info.IsDir() || !match(path, config.MatchRegexes) {
			return nil
		}

		// Append data file.
		data, err := util.ReadFile(repoFS, path)
		if err != nil {
			return fmt.Errorf("could not read git file: %w", err)
		}
		files = append(files, string(data))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk git repository: %w", err)
	}

	return storage.StaticSourceCodeRepository(files), nil
}

func match(path string, rs []*regexp.Regexp) bool {
	for _, r := range rs {
		if r.MatchString(path) {
			return true
		}
	}

	return false
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

		_, err = git.Clone(storer, memfs, &git.CloneOptions{
			URL:           config.URL,
			Depth:         1,
			ReferenceName: ref,
		})
		if err == nil {
			// Store in cache.
			pluginFsCache[id] = memfs
			return memfs, nil
		}
	}

	return nil, fmt.Errorf("could not clone repository: %w", err)
}
