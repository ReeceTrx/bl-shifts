package store

import "context"

type Store interface {
	FilterAndSaveCodes(ctx context.Context, codes []string) ([]string, error)
}
