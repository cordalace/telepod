package versionsdb

import (
	"context"
	"encoding/json"
	"errors"
	"os"

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
	var db string
	var err error

	if os.Geteuid() == 0 {
		dir := "/var/lib/telepod"
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				if err2 := os.MkdirAll(dir, os.ModePerm); err2 != nil {
					return "", err2
				}
			} else {
				return "", err
			}
		}
		db = "/var/lib/telepod/db.json"
	} else {
		db, err = xdg.StateFile(appName + "/db.json")
		if err != nil {
			return "", err
		}
	}

	return db, nil
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
		return err
	}

	if err := json.Unmarshal(data, &d.cfg); err != nil {
		return err
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

func (d *VersionsDB) GetContainer(ctx context.Context, name string) (*workflow.Container, error) {
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

func (d *VersionsDB) CreateContainer(ctx context.Context, container *workflow.Container) error {
	newContainer := &containerJSON{Name: container.Name, Version: container.ImageVersion}
	d.cfg.Containers = append(d.cfg.Containers, newContainer)
	d.containersByName[newContainer.Name] = newContainer

	return nil
}

func (d *VersionsDB) UpdateContainer(ctx context.Context, container *workflow.Container) error {
	c := d.getContainer(container.Name)
	if c == nil {
		return workflow.ErrContainerNotFound
	}

	c.Version = container.ImageVersion

	return nil
}

func (d *VersionsDB) Flush(ctx context.Context) error {
	db, err := d.dbPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(d.cfg, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(db, data, 0o600)
}
