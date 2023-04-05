package workflow

import (
	"context"
	"errors"
	"fmt"
)

func (w *Workflow) Run(ctx context.Context) error {
	runningContainers, err := w.podRuntime.ListRunningContainers(ctx)
	if err != nil {
		return fmt.Errorf("error listing containers from database: %w", err)
	}

	isVersionsDBChanged := false
	for _, runningContainer := range runningContainers {
		dbContainer, containerFirstSeen, err := w.ensureContainerInDatabase(ctx, runningContainer)
		if err != nil {
			return err
		}

		if containerFirstSeen {
			isVersionsDBChanged = true

			continue
		}

		if dbContainer.ImageVersion != runningContainer.ImageVersion {
			if err := w.handleNewImage(ctx, runningContainer); err != nil {
				return err
			}

			isVersionsDBChanged = true
		}
	}

	if isVersionsDBChanged {
		if err := w.versionsDB.Flush(ctx); err != nil {
			return fmt.Errorf("error flushing database: %w", err)
		}
	}

	return nil
}

func (w *Workflow) ensureContainerInDatabase(
	ctx context.Context,
	runningContainer *Container,
) (*Container, bool, error) {
	containerFirstSeen := false
	dbContainer, err := w.versionsDB.GetContainer(ctx, runningContainer.Name)
	switch {
	case errors.Is(err, ErrContainerNotFound):
		if err := w.versionsDB.CreateContainer(ctx, runningContainer); err != nil {
			return nil, false, fmt.Errorf("error creating container in database: %w", err)
		}
		containerFirstSeen = true
	case err != nil:
		return nil, false, fmt.Errorf("error receiving container from database: %w", err)
	default:
		// nop
	}

	return dbContainer, containerFirstSeen, nil
}

func (w *Workflow) handleNewImage(ctx context.Context, container *Container) error {
	if err := w.versionsDB.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("error updating container in database: %w", err)
	}

	if err := w.notifier.CreateNotification(ctx, container); err != nil {
		return fmt.Errorf("error sending notification: %w", err)
	}

	return nil
}
