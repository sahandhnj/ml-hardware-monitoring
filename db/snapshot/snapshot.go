package snapshot

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	boltService "github.com/sahandhnj/ml-hardware-monitoring/pkg/bolt"
	"github.com/sahandhnj/ml-hardware-monitoring/types"
)

const (
	BucketName = "snapshot"
)

type Service struct {
	db *bolt.DB
}

func NewService(db *bolt.DB) (*Service, error) {
	err := boltService.CreateBucket(db, BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		db: db,
	}, nil
}

func (s *Service) Snapshot(ID int) (*types.SnapShot, error) {
	var snapshot types.SnapShot
	identifier := boltService.Itob(int(ID))

	err := boltService.GetObject(s.db, BucketName, identifier, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (s *Service) SnapshotByDeviceUUID(uuid string) (*types.SnapShot, error) {
	var snapshot *types.SnapShot

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var ss types.SnapShot
			err := json.Unmarshal(v, &ss)
			if err != nil {
				return err
			}

			if ss.DeviceUUID == uuid {
				snapshot = &ss
				break
			}
		}

		if snapshot == nil {
			return boltService.GetError("not found")
		}

		return nil
	})

	return snapshot, err
}

func (s *Service) Snapshots() ([]types.SnapShot, error) {
	var snapshots = make([]types.SnapShot, 0)

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var snapshot types.SnapShot
			err := json.Unmarshal(v, &snapshot)
			if err != nil {
				return err
			}
			snapshots = append(snapshots, snapshot)
		}

		return nil
	})

	return snapshots, err
}

func (s *Service) GetNextIdentifier() int {
	return boltService.GetNextIdentifier(s.db, BucketName)
}

func (s *Service) CreateSnapshot(snapshot *types.SnapShot) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		err := bucket.SetSequence(uint64(snapshot.ID))
		if err != nil {
			return err
		}

		data, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}

		return bucket.Put(boltService.Itob(int(snapshot.ID)), data)
	})
}

func (s *Service) UpdateSnapshot(ID int, snapshot *types.SnapShot) error {
	identifier := boltService.Itob(int(ID))
	return boltService.UpdateObject(s.db, BucketName, identifier, snapshot)
}

func (s *Service) DeleteSnapshot(ID int) error {
	identifier := boltService.Itob(int(ID))
	return boltService.DeleteObject(s.db, BucketName, identifier)
}
