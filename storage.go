// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	// "log"
	// "fmt"

	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

const (
	dbBucket = "2faKeys"
)

type Storage struct {
	db       *bolt.DB
	filename string
	folder   string
}

func NewStorage(folder string, filename string) Storage {
	storage := Storage{
		folder:   folder,
		filename: filename,
	}

	return storage
}

func (s *Storage) Init() error {
	db, err := bolt.Open(filepath.Join(s.folder, s.filename), 0600, nil)
	if err != nil {
		return errors.Errorf("open database: %q", err)
	}
	s.db = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dbBucket))
		if err != nil {
			return errors.Errorf("create bucket: %q", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) AddKey(key string, value []byte) (bool, error) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbBucket))

		err := bucket.Put([]byte(key), value)
		if err != nil {
			return errors.Errorf("can't put: %q", err)
		}

		return nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Storage) ListKey() ([]string, error) {
	var keys []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			keys = append(keys, string(k))
		}

		return nil
	})
	if err != nil {
		return []string{}, err
	}

	return keys, nil
}

func (s *Storage) GetKey(key string) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		bucket := tx.Bucket([]byte(dbBucket))
		if bucket == nil {
			return errors.Errorf("Bucket %q not found!", dbBucket)
		}

		value = bucket.Get([]byte(key))

		return nil
	})
	if err != nil {
		return []byte{}, err
	}

	return value, nil
}

func (s *Storage) RemoveKey(key string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		bucket := tx.Bucket([]byte(dbBucket))
		if bucket == nil {
			return errors.Errorf("Bucket %q not found!", dbBucket)
		}

		err := bucket.Delete([]byte(key))
		if err != nil {
			return errors.Errorf("can't delete %q: %q", key, err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
