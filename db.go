package main

import (
	"crypto/sha1"
	"errors"
	errors2 "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"time"
)

var (
	globalDB *bolt.DB

	bucket = []byte("news")

	Err_not_found = errors.New("key not found")
)

func openDB(path string) {
	db, err := bolt.Open(path, 0660, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
		return
	}
	globalDB = db

	err = GetDB().Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return
	}

}

func GetDB() *bolt.DB {
	return globalDB
}

func Save(key, value string) error {
	err := GetDB().Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(bucket).Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors2.WithStack(err)
	}

	return nil
}

func Find(key string) (string, error) {
	var r string

	err := GetDB().View(func(tx *bolt.Tx) error {
		bt := tx.Bucket(bucket).Get([]byte(key))
		if bt == nil {
			return Err_not_found
		}
		r = string(bt)
		return nil
	})
	if err != nil {
		return r, errors2.WithStack(err)
	}

	return r, nil
}

func hashKey(str string) string {
	h := sha1.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		log.Fatalf("err: %+v", errors2.WithStack(err))
		return ""
	}
	return string(h.Sum(nil))
}
