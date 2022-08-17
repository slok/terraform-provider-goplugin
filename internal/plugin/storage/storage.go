package storage

import "context"

type SourceCodeRepository interface {
	GetSourceCode(ctx context.Context) ([]string, error)
}

// DataSourceCodeRepository will be used to make a []string behave like a source code repository.
type DataSourceCodeRepository []string

func (d DataSourceCodeRepository) GetSourceCode(ctx context.Context) ([]string, error) {
	return d, nil
}
