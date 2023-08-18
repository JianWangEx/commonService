package manager

import "context"

// @Injector()
type GameManager interface {
	Find(ctx context.Context, name string) (string, error)
}
