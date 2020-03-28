package datastore

import (
	"context"

	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
)

// KVStore is the key value datastore
type KVStore struct {
	db *badger.DB
}

// NewKVStore creates a new key value datastore
func NewKVStore(path string) (*KVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open datastore")
	}
	return &KVStore{
		db,
	}, nil
}

// Get retrieves key values from the datastore.
func (s *KVStore) Get(request *apiv1.GetValuesRequest) (*apiv1.GetValuesResponse, error) {
	response := &apiv1.GetValuesResponse{
		Messages: []*apiv1.KeyValue{},
	}

	err := s.db.View(func(txn *badger.Txn) error {
		for _, key := range request.Requests {
			if key.IsPrefix {
				values, err := s.prefixScan(txn, key.Key)
				if err != nil {
					return errors.Wrap(err, "failed prefix scan")
				}
				response.Messages = append(response.Messages, values...)
				continue
			}
			value, err := txn.Get([]byte(key.Key))
			if err == badger.ErrKeyNotFound {
				continue
			}
			if err != nil {
				return errors.Wrap(err, "failed to get key")
			}

			err = value.Value(func(v []byte) error {
				response.Messages = append(response.Messages, &apiv1.KeyValue{
					Key:   string(key.Key),
					Value: v,
				})
				return nil
			})
			if err != nil {
				return errors.Wrap(err, "failed to get value")
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get from db")
	}

	return response, nil
}

// Set sets key values in the datastore.
func (s *KVStore) Set(request *apiv1.SetValuesRequest) error {
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, value := range request.Messages {
		err := wb.Set([]byte(value.Key), []byte(value.Value))
		if err != nil {
			return errors.Wrap(err, "failed to set key")
		}
	}

	err := wb.Flush()
	if err != nil {
		return errors.Wrap(err, "failed to flush")
	}

	return nil
}

// Subscribe will subscribe to prefixes in the key value store. This will block until there is an error
// or the context is cancelled
func (s *KVStore) Subscribe(ctx context.Context, subscription *apiv1.SubscribeRequest, cb func(*apiv1.SubscribeResponse) error) error {
	err := s.db.View(func(txn *badger.Txn) error {
		for _, key := range subscription.Prefixes {
			values, err := s.prefixScan(txn, key)
			if err != nil {
				return errors.Wrap(err, "failed prefix scan")
			}

			err = cb(&apiv1.SubscribeResponse{
				Messages: values,
			})
			if err != nil {
				return errors.Wrap(err, "failed callback")
			}

			return nil
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to get")
	}

	prefixes := [][]byte{}
	for _, p := range subscription.Prefixes {
		prefixes = append(prefixes, []byte(p))
	}

	err = s.db.Subscribe(ctx, func(kv *badger.KVList) error {
		values := []*apiv1.KeyValue{}
		for _, kv := range kv.Kv {
			values = append(values, &apiv1.KeyValue{
				Key:   string(kv.Key),
				Value: kv.Value,
			})
		}

		err := cb(&apiv1.SubscribeResponse{
			Messages: values,
		})
		if err != nil {
			return errors.Wrap(err, "failed callback")
		}

		return nil
	}, prefixes...)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}

	return nil
}

func (s *KVStore) prefixScan(txn *badger.Txn, prefixKey string) ([]*apiv1.KeyValue, error) {
	values := []*apiv1.KeyValue{}

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte(prefixKey)
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := item.Key()
		err := item.Value(func(v []byte) error {
			values = append(values, &apiv1.KeyValue{
				Key:   string(k),
				Value: v,
			})
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value")
		}
	}

	return values, nil
}