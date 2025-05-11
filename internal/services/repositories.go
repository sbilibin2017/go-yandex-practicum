package services

import "context"

type FilterOneRepository interface {
	FilterOne(ctx context.Context, filter map[string]any) (map[string]any, error)
}

type SaveRepository interface {
	Save(ctx context.Context, data map[string]any) error
}
