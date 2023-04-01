package workflow

import (
	"context"
	"errors"
)

var ErrContainerNotFound = errors.New("container not found")

type VersionsDB interface {
	GetContainer(ctx context.Context, name string) (*Container, error)
	CreateContainer(ctx context.Context, container *Container) error
	UpdateContainer(ctx context.Context, container *Container) error
	Flush(ctx context.Context) error
}
