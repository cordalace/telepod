package podruntime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"codeberg.org/cordalace/telepod/internal/workflow"
)

const (
	podmanPath = "/usr/bin/podman"
)

func NewPodRuntime() *PodRuntime {
	return &PodRuntime{}
}

type PodRuntime struct{}

func (r *PodRuntime) ListRunningContainers(ctx context.Context) ([]*workflow.Container, error) {
	containerNames, err := r.listContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}

	ret := make([]*workflow.Container, len(containerNames))
	for idx, containerName := range containerNames {
		imageID, err := r.getImageID(ctx, containerName)
		if err != nil {
			return nil, fmt.Errorf("error finding container image id: %s: %w", containerName, err)
		}

		buildVersion, err := r.getBuildVersion(ctx, imageID)
		if err != nil {
			return nil, fmt.Errorf("error finding container image build version: %s: %w", imageID, err)
		}

		ret[idx] = &workflow.Container{Name: containerName, ImageVersion: buildVersion}
	}

	return ret, nil
}

func (r *PodRuntime) podman(ctx context.Context, cmd ...string) (string, error) {
	c := exec.CommandContext(ctx, podmanPath, cmd...)
	var out bytes.Buffer
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("error running podman command: %s: %w", r.formatShell(cmd), err)
	}

	return strings.TrimSpace(out.String()), nil
}

func (r *PodRuntime) formatShell(cmd []string) string {
	return podmanPath + " " + strings.Join(cmd, " ")
}

func (r *PodRuntime) getImageID(ctx context.Context, container string) (string, error) {
	imageID, err := r.podman(ctx, "inspect", "--format", "{{ .Image }}", container)
	if err != nil {
		return "", err
	}

	return imageID, nil
}

func (r *PodRuntime) getBuildVersion(ctx context.Context, imageID string) (string, error) {
	buildVersion, err := r.podman(
		ctx,
		"inspect",
		"--format",
		"{{index .Labels \"org.opencontainers.image.version\"}}",
		imageID,
	)
	if err != nil {
		return "", err
	}

	return buildVersion, nil
}

func (r *PodRuntime) listContainers(ctx context.Context) ([]string, error) {
	lines, err := r.podman(ctx, "ps", "--format", "{{ .Names }}")
	if err != nil {
		return nil, err
	}

	if lines == "" {
		return nil, nil
	}

	return strings.Split(lines, "\n"), nil
}
