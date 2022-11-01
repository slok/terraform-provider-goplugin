package storage

import (
	"context"
	"io/fs"
)

type SourceCodeRepository interface {
	FS(ctx context.Context) fs.FS
	Index(ctx context.Context) string
	Gopath(ctx context.Context) string
	ImportPath(ctx context.Context) string
}
