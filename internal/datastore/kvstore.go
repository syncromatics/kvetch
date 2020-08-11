package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
)

//KVStoreOptions represent environment variable configurable options related to the KV Store.
type KVStoreOptions struct {
	EnableTruncate                              *wrappers.BoolValue
	MaxTableSize                                *wrappers.Int64Value
	LevelOneSize                                *wrappers.Int64Value
	LevelSizeMultiplier                         *wrappers.Int32Value
	NumberOfLevelZeroTables                     *wrappers.Int32Value
	NumberOfLevelZeroTablesUntilForceCompaction *wrappers.Int32Value
	GarbageCollectionDiscardRatio               *wrappers.FloatValue
}

// KVStore is the key value datastore
type KVStore struct {
	db                            *badger.DB
	garbageCollectionDiscardRatio float64
}

func getBadgerOptions(path string, options *KVStoreOptions) badger.Options {
	opts := badger.DefaultOptions(path)
	if options.EnableTruncate != nil {
		fmt.Printf("Configuring with enableTruncate: %t \n", options.EnableTruncate.Value)
		opts = opts.WithTruncate(options.EnableTruncate.Value)
	}
	if options.MaxTableSize != nil {
		fmt.Printf("Configuring with maxTableSize: %d \n", options.MaxTableSize.Value)
		opts = opts.WithMaxTableSize(options.MaxTableSize.Value)
	}
	if options.LevelOneSize != nil {
		fmt.Printf("Configuring with LevelOneSize: %d \n", options.LevelOneSize.Value)
		opts = opts.WithLevelOneSize(options.LevelOneSize.Value)
	}
	if options.LevelSizeMultiplier != nil {
		fmt.Printf("Configuring with LevelSizeMultiplier: %d \n", options.LevelSizeMultiplier.Value)
		opts = opts.WithLevelSizeMultiplier(int(options.LevelSizeMultiplier.Value))
	}
	if options.NumberOfLevelZeroTables != nil {
		fmt.Printf("Configuring with NumberOfLevelZeroTables: %d \n", options.NumberOfLevelZeroTables.Value)
		opts = opts.WithNumLevelZeroTables(int(options.NumberOfLevelZeroTables.Value))
	}
	if options.NumberOfLevelZeroTablesUntilForceCompaction != nil {
		fmt.Printf("Configuring with NumberOfLevelZeroTablesUntilForceCompaction: %d \n", options.NumberOfLevelZeroTablesUntilForceCompaction.Value)
		opts = opts.WithNumLevelZeroTablesStall(int(options.NumberOfLevelZeroTablesUntilForceCompaction.Value))
	}

	return opts
}

// NewKVStore creates a new key value datastore
func NewKVStore(path string, options *KVStoreOptions) (*KVStore, error) {
	garbageCollectionDiscardRatio := 0.5
	if options.GarbageCollectionDiscardRatio != nil {
		fmt.Printf("Configuring with garbageCollectionDiscardRatio: %f \n", options.GarbageCollectionDiscardRatio.Value)
		garbageCollectionDiscardRatio = float64(options.GarbageCollectionDiscardRatio.Value)
	}

	opts := getBadgerOptions(path, options)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open datastore")
	}
	return &KVStore{
		db,
		garbageCollectionDiscardRatio,
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
					Value: append([]byte{}, v...),
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

	if request.TtlDuration != nil {
		ttl, err := ptypes.Duration(request.TtlDuration)
		if err != nil {
			return errors.Wrap(err, "failed to deserialize ttl")
		}
		expire := uint64(time.Now().Add(ttl).Unix())
		for _, value := range request.Messages {
			entry := &badger.Entry{
				Key:       []byte(value.Key),
				Value:     []byte(value.Value),
				ExpiresAt: expire,
			}
			err = wb.SetEntry(entry)
			if err != nil {
				return errors.Wrap(err, "failed to set key")
			}
		}
	} else {
		for _, value := range request.Messages {
			err := wb.Set([]byte(value.Key), []byte(value.Value))
			if err != nil {
				return errors.Wrap(err, "failed to set key")
			}
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

// GarbageCollect cleans up old values in log files
func (s *KVStore) GarbageCollect() error {
	err := s.db.RunValueLogGC(s.garbageCollectionDiscardRatio)
	if err == badger.ErrNoRewrite { // no cleanup happened, this is okay
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed log gc")
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
				Value: append([]byte{}, v...),
			})
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get value")
		}
	}

	return values, nil
}
