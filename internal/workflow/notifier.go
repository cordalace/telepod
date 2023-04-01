package workflow

import "context"

type Notifier interface {
	CreateNotification(ctx context.Context, container *Container) error
}
