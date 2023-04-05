package versionsdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"

	"codeberg.org/cordalace/telepod/internal/workflow"
	"github.com/adrg/xdg"
)

const (
	databaseVersion = 1
	appName         = "telepod"
)

var errUnsupported = errors.New("unsupported versions database version")

func NewVersionsDB() *VersionsDB {
	return &VersionsDB{cfg: nil, containersByName: nil}
}

type VersionsDB struct {
	cfg              *configJSON
	containersByName map[string]*containerJSON
}

type configJSON struct {
	Containers []*containerJSON `json:"containers"`
	Version    int              `json:"version"`
}

type containerJSON struct {
	Name    string `json:"name"`
	Version string `json:"tag"`
}

func (d *VersionsDB) dbPath() (string, error) {
	var dbLocation string
	var err error

	if os.Geteuid() == 0 {
		if err := d.ensureDir("/var/lib/telepod"); err != nil {
			return "", err
		}
		dbLocation = "/var/lib/telepod/db.json"
	} else {
		dbLocation, err = xdg.StateFile(appName + "/db.json")
		if err != nil {
			return "", fmt.Errorf("error generating json database location: %w", err)
		}
	}

	return dbLocation, nil
}

func (d *VersionsDB) ensureDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err2 := os.MkdirAll(dir, os.ModePerm); err2 != nil {
				return fmt.Errorf("error creating json database directory: %w", err2)
			}
		} else {
			return fmt.Errorf("error accessing json database directory: %w", err)
		}
	}

	return nil
}

func (d *VersionsDB) Init() error {
	db, err := d.dbPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(db)
	if err != nil {
		if os.IsNotExist(err) {
			d.cfg = &configJSON{
				Containers: []*containerJSON{},
				Version:    databaseVersion,
			}
			d.containersByName = make(map[string]*containerJSON, 0)

			return nil
		}

		return fmt.Errorf("error reading json database file: %w", err)
	}

	if err := json.Unmarshal(data, &d.cfg); err != nil {
		return fmt.Errorf("error unmarshaling json database: %w", err)
	}

	if d.cfg.Version == 0 {
		d.cfg.Version = databaseVersion
	}

	if d.cfg.Version > databaseVersion {
		return errUnsupported
	}

	d.containersByName = make(map[string]*containerJSON, len(d.cfg.Containers))
	for _, container := range d.cfg.Containers {
		d.containersByName[container.Name] = container
	}

	return nil
}

func (d *VersionsDB) GetContainer(_ context.Context, name string) (*workflow.Container, error) {
	container := d.getContainer(name)
	if container == nil {
		return nil, workflow.ErrContainerNotFound
	}

	return &workflow.Container{Name: container.Name, ImageVersion: container.Version}, nil
}

func (d *VersionsDB) getContainer(name string) *containerJSON {
	container, ok := d.containersByName[name]
	if !ok {
		return nil
	}

	return container
}

func (d *VersionsDB) CreateContainer(_ context.Context, container *workflow.Container) error {
	newContainer := &containerJSON{Name: container.Name, Version: container.ImageVersion}
	d.cfg.Containers = append(d.cfg.Containers, newContainer)
	d.containersByName[newContainer.Name] = newContainer

	return nil
}

func (d *VersionsDB) UpdateContainer(_ context.Context, container *workflow.Container) error {
	c := d.getContainer(container.Name)
	if c == nil {
		return workflow.ErrContainerNotFound
	}

	c.Version = container.ImageVersion

	return nil
}

func (d *VersionsDB) Flush(_ context.Context) error {
	dbLocation, err := d.dbPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(d.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling json database: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(dbLocation, data, syscall.S_IRUSR|syscall.S_IWUSR); err != nil {
		return fmt.Errorf("error writing json database file: %w", err)
	}

	return nil
}
