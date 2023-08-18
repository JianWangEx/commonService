package handler

import "context"

// @Injector()
type SpnHandler interface {
	Deal(ctx context.Context, name string) (bool, error)
}
