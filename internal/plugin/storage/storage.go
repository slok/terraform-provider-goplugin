package storage

import "context"

type SourceCodeRepository interface {
	GetSourceCode(ctx context.Context) ([]string, error)
}

// StaticSourceCodeRepository will be used to make a []string behave like a source code repository.
type StaticSourceCodeRepository []string

func (s StaticSourceCodeRepository) GetSourceCode(ctx context.Context) ([]string, error) {
	return s, nil
}
