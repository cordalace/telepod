package workflow

import (
	"context"
	"errors"
)

func (w *Workflow) Run(ctx context.Context) error {
	runningContainers, err := w.podRuntime.ListRunningContainers(ctx)
	if err != nil {
		return err
	}

	isVersionsDBChanged := false
	for _, runningContainer := range runningContainers {
		cachedContainer, err := w.versionsDB.GetContainer(ctx, runningContainer.Name)
		if err != nil {
			if errors.Is(err, ErrContainerNotFound) {
				if err := w.versionsDB.CreateContainer(ctx, runningContainer); err != nil {
					return err
				}
				isVersionsDBChanged = true

				continue
			}

			return err
		}

		if cachedContainer.ImageVersion != runningContainer.ImageVersion {
			if err := w.versionsDB.UpdateContainer(ctx, runningContainer); err != nil {
				return err
			}

			if err := w.notifier.CreateNotification(ctx, runningContainer); err != nil {
				return err
			}

			isVersionsDBChanged = true
		}
	}

	if isVersionsDBChanged {
		if err := w.versionsDB.Flush(ctx); err != nil {
			return err
		}
	}

	return nil
}
