package manager

import "context"

// @Injector()
type HunterManager interface {
	Find(ctx context.Context, name string) (string, error)
}
