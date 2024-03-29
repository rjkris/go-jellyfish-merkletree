package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"sync"
)

type KVDatabase struct {
	Db       *leveldb.DB
	fn       string
	quitLock sync.Mutex
}

func New(file string) (*KVDatabase, error) {
	db, err := leveldb.OpenFile(file, nil)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	if err != nil {
		return nil, err
	}
	kvdb := &KVDatabase{
		Db:       db,
		fn:       file,
		quitLock: sync.Mutex{},
	}
	return kvdb, nil
}

func (kvdb *KVDatabase)Close() error {
	kvdb.quitLock.Lock()
	defer kvdb.quitLock.Unlock()
	return kvdb.Db.Close()
}

func (kvdb *KVDatabase)Get(key []byte) ([]byte, error) {
	res, err := kvdb.Db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (kvdb *KVDatabase)Put(key []byte, value []byte) error {
	return kvdb.Db.Put(key, value, nil)
}
