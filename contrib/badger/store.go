package badger

import (
	"fmt"

	"github.com/dlshle/aghs/store"
	"github.com/dlshle/aghs/utils"

	badger "github.com/dgraph-io/badger/v3"
)

type BadgerStoreSerializeHandler interface {
	KeySerializer(interface{}) ([]byte, error)
	KeyDeserializer([]byte) (interface{}, error)
	ValueSerializer(interface{}) ([]byte, error)
	ValueDeserializer([]byte) (interface{}, error)
}

type BadgerStore struct {
	db *badger.DB
	BadgerStoreSerializeHandler
}

func NewBadgerStore(dbFilePath string, serializeHandler BadgerStoreSerializeHandler) (store.AdvancedKVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(dbFilePath))
	if err != nil {
		return nil, err
	}
	return BadgerStore{
		db:                          db,
		BadgerStoreSerializeHandler: serializeHandler,
	}, nil
}

func (s BadgerStore) withRead(cb func(tx *badger.Txn) error) error {
	return s.db.View(cb)
}

func (s BadgerStore) withWrite(cb func(tx *badger.Txn) error) error {
	return s.db.Update(cb)
}

func (s BadgerStore) Get(key interface{}) (res interface{}, err error) {
	err = s.withRead(func(tx *badger.Txn) error {
		serializedKey, err := s.KeySerializer(key)
		if err != nil {
			return err
		}
		res, err = s.getValueBySerializedKey(tx, serializedKey)
		return err
	})
	return
}

func (s BadgerStore) getValueBySerializedKey(tx *badger.Txn, key []byte) (value interface{}, err error) {
	var item *badger.Item
	err = utils.ProcessWithErrors(func() error {
		item, err = tx.Get(key)
		if err == badger.ErrKeyNotFound {
			// badger returns ErrKeyNotFound on record DNE
			err = nil
		}
		return err
	}, func() error {
		return item.Value(func(val []byte) error {
			if item == nil {
				return nil
			}
			value, err = s.ValueDeserializer(val)
			return err
		})
	})
	return
}

func (s BadgerStore) Has(key interface{}) (bool, error) {
	val, err := s.Get(key)
	return val == nil, err
}

func (s BadgerStore) Put(key interface{}, value interface{}) (success bool, err error) {
	err = s.withWrite(func(tx *badger.Txn) error {
		serializedKey, serializedValue, err := s.serializeKV(key, value)
		if err != nil {
			return err
		}
		return tx.Set(serializedKey, serializedValue)
	})
	return err == nil, err
}

func (s BadgerStore) Update(key interface{}, value interface{}) (success bool, err error) {
	err = s.withWrite(func(tx *badger.Txn) error {
		serializedKey, serializedValue, err := s.serializeKV(key, value)
		if err != nil {
			return err
		}
		_, err = tx.Get(serializedKey)
		// update only when record exists
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("record %s does not exist", string(serializedKey))
		}
		return tx.Set(serializedKey, serializedValue)
	})
	return err == nil, err
}

func (s BadgerStore) Delete(key interface{}) (bool, error) {
	err := s.withWrite(func(tx *badger.Txn) error {
		serializedKey, err := s.KeySerializer(key)
		if err != nil {
			return err
		}
		return tx.Delete(serializedKey)
	})
	return err != nil, err
}

func (s BadgerStore) Query(filter func(record interface{}) bool) (res []interface{}, err error) {
	err = s.iterate(func(_, value interface{}) error {
		if filter(value) {
			res = append(res, value)
		}
		return nil
	})
	return
}

func (s BadgerStore) BulkGet(keys []interface{}) (res []interface{}, err error) {
	keySet := make(map[interface{}]bool)
	for _, key := range keys {
		keySet[key] = true
	}
	err = s.iterate(func(key, value interface{}) error {
		if keySet[key] {
			res = append(res, value)
		}
		return nil
	})
	return
}

func (s BadgerStore) BulkPut(bulk map[interface{}]interface{}) (success bool, err error) {
	err = s.withWrite(func(tx *badger.Txn) error {
		for key, value := range bulk {
			var serializedKey, serializedValue []byte
			utils.ProcessWithErrors(func() error {
				serializedKey, serializedValue, err = s.serializeKV(key, value)
				return err
			}, func() error {
				return tx.Set(serializedKey, serializedValue)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (s BadgerStore) iterate(cb func(k interface{}, v interface{}) error) error {
	return s.withRead(func(tx *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		itr := tx.NewIterator(opt)
		defer itr.Close()
		for itr.Rewind(); itr.Valid(); itr.Next() {
			item := itr.Item()
			rawKey := item.Key()
			var key, value interface{}
			var err error
			item.Value(func(val []byte) error {
				key, value, err = s.deserializeKV(rawKey, val)
				return err
			})
			err = cb(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s BadgerStore) Close() error {
	return s.db.Close()
}

func (s BadgerStore) serializeKV(key interface{}, value interface{}) (k, v []byte, e error) {
	e = utils.ProcessWithErrors(func() error {
		if key != nil {
			k, e = s.KeySerializer(key)
		}
		return e
	}, func() error {
		if value != nil {
			v, e = s.ValueSerializer(value)
		}
		return e
	})
	return
}

func (s BadgerStore) deserializeKV(rawKey []byte, rawValue []byte) (key, value interface{}, e error) {
	e = utils.ProcessWithErrors(func() error {
		if rawKey != nil {
			key, e = s.KeyDeserializer(rawKey)
		}
		return e
	}, func() error {
		if rawValue != nil {
			value, e = s.ValueDeserializer(rawValue)
		}
		return e
	})
	return
}
