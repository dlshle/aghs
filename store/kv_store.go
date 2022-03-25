package store

type KVStore interface {
	Get(key interface{}) (interface{}, error)
	Has(key interface{}) (bool, error)
	Put(key interface{}, value interface{}) (bool, error) // insert if dne, update if exists
	Update(key interface{}, value interface{}) (bool, error)
	Delete(key interface{}) (bool, error)
}

type AdvancedKVStore interface {
	KVStore
	Query(filter func(record interface{}) bool) ([]interface{}, error)
	BulkGet(keys []interface{}) ([]interface{}, error)
	BulkPut(bulk map[interface{}]interface{}) (bool, error)
}
