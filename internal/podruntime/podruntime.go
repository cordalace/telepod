package podruntime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"codeberg.org/cordalace/telepod/internal/workflow"
)

func NewPodRuntime() *PodRuntime {
	return &PodRuntime{}
}

type PodRuntime struct {
	podmanPath string
}

func (r *PodRuntime) Init() error {
	var err error

	r.podmanPath, err = exec.LookPath("podman")
	if err != nil {
		return fmt.Errorf("could not find podman binary in PATH: %w", err)
	}

	return nil
}

func (r *PodRuntime) ListRunningContainers(ctx context.Context) ([]*workflow.Container, error) {
	containerNames, err := r.listContainers()
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}

	ret := make([]*workflow.Container, len(containerNames))
	for i, containerName := range containerNames {
		imageID, err := r.getImageID(containerName)
		if err != nil {
			return nil, fmt.Errorf("error finding container image id: %s: %w", containerName, err)
		}

		buildVersion, err := r.getBuildVersion(imageID)
		if err != nil {
			return nil, fmt.Errorf("error finding container image build version: %s: %w", imageID, err)
		}

		ret[i] = &workflow.Container{Name: containerName, ImageVersion: buildVersion}
	}

	return ret, nil
}

func (r *PodRuntime) podman(cmd ...string) (string, error) {
	// should fix subprocess via variable
	c := exec.Command(r.podmanPath, cmd...) // nosec
	var out bytes.Buffer
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func (r *PodRuntime) getImageID(container string) (string, error) {
	imageID, err := r.podman("inspect", "--format", "{{ .Image }}", container)
	if err != nil {
		return "", err
	}

	return imageID, nil
}

func (r *PodRuntime) getBuildVersion(imageID string) (string, error) {
	buildVersion, err := r.podman("inspect", "--format", "{{index .Labels \"org.opencontainers.image.version\"}}", imageID)
	if err != nil {
		return "", err
	}

	return buildVersion, nil
}

func (r *PodRuntime) listContainers() ([]string, error) {
	lines, err := r.podman("ps", "--format", "{{ .Names }}")
	if err != nil {
		return nil, err
	}

	if lines == "" {
		return nil, nil
	}

	return strings.Split(lines, "\n"), nil
}
