package workflow

import "context"

type PodRuntime interface {
	ListRunningContainers(ctx context.Context) ([]*Container, error)
}
