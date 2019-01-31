package db

import (
	"os"
	"os/user"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sahandhnj/ml-hardware-monitoring/db/snapshot"
)

const (
	databaseFileName = "hwop.db"
	StorePath        = ".hwop"
	TIMEOUT_SECONDS  = 1
)

type DBService struct {
	path            string
	db              *bolt.DB
	SnapShotService *snapshot.Service
}

func NewDBService() (*DBService, error) {
	usr, err := user.Current()
	databaseDir := path.Join(usr.HomeDir, StorePath)
	databasePath := path.Join(databaseDir, databaseFileName)

	if _, err := os.Stat(databaseDir); os.IsNotExist(err) {
		err = os.Mkdir(databaseDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	DBService := &DBService{
		path: databasePath,
	}

	err = DBService.Open()
	if err != nil {
		return nil, err
	}

	err = DBService.initServices()
	if err != nil {
		return nil, err
	}

	return DBService, nil
}

func (d *DBService) Open() error {
	db, err := bolt.Open(d.path, 0600, &bolt.Options{Timeout: TIMEOUT_SECONDS * time.Second})
	if err != nil {
		return err
	}
	d.db = db

	return d.initServices()
}

func (d *DBService) Close() error {
	if d.db != nil {
		return d.db.Close()
	}

	return nil
}

func (d *DBService) initServices() error {
	snapShotServie, err := snapshot.NewService(d.db)
	if err != nil {
		return err
	}

	d.SnapShotService = snapShotServie

	return nil
}
