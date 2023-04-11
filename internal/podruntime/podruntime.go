package podruntime

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"codeberg.org/cordalace/telepod/internal/workflow"
	"github.com/adrg/xdg"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/images"
)

func NewPodRuntime() *PodRuntime {
	return &PodRuntime{
		conn: nil,
	}
}

type PodRuntime struct {
	conn context.Context //nolint:containedctx
}

func (r *PodRuntime) Init() error {
	endpoint, err := getDefaultEndpoint()
	if err != nil {
		return err
	}

	r.conn, err = bindings.NewConnection(context.Background(), endpoint)
	if err != nil {
		return fmt.Errorf("error creating connection: %w", err)
	}

	return nil
}

func getDefaultEndpoint() (string, error) {
	// The default endpoint for a rootful service is unix:///run/podman/podman.sock
	// and rootless is unix://$XDG_RUNTIME_DIR/podman/podman.sock
	var path string
	var err error

	if os.Geteuid() == 0 {
		path = "/run/podman/podman.sock"
	} else {
		path, err = xdg.RuntimeFile("podman/podman.sock")
		if err != nil {
			return "", fmt.Errorf("error generating podman.sock location: %w", err)
		}
	}

	endpointURL := url.URL{
		Scheme:      "unix",
		Opaque:      "",
		User:        nil,
		Host:        "",
		Path:        path,
		RawPath:     "",
		OmitHost:    false,
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	return endpointURL.String(), nil
}

func (r *PodRuntime) ListRunningContainers(_ context.Context) ([]*workflow.Container, error) {
	containers, err := containers.List(r.conn, nil) //nolint:contextcheck
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}

	ret := make([]*workflow.Container, len(containers))
	for idx, container := range containers {
		// TODO: pick all names
		containerName := container.Names[0]
		img, err := images.GetImage(r.conn, container.ImageID, nil) //nolint:contextcheck
		if err != nil {
			return nil, fmt.Errorf("error getting container image id: %s: %w", containerName, err)
		}
		buildVersion := img.Labels["org.opencontainers.image.version"]
		ret[idx] = &workflow.Container{Name: containerName, ImageVersion: buildVersion}
	}

	return ret, nil
}
