package datastore_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/syncromatics/kvetch/internal/datastore"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"

	"gotest.tools/assert"
)

func Test_GetPrefix(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "Test_GetPrefix")
	assert.NilError(t, err)

	store, err := datastore.NewKVStore(tmpDir)
	assert.NilError(t, err)

	err = store.Set(&apiv1.SetValuesRequest{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2"),
			},
			&apiv1.KeyValue{
				Key:   "test/2/stuff",
				Value: []byte("bad value"),
			},
		},
	})
	assert.NilError(t, err)

	values, err := store.Get(&apiv1.GetValuesRequest{
		Requests: []*apiv1.GetValuesRequest_GetValue{
			&apiv1.GetValuesRequest_GetValue{
				Key:      "test/1",
				IsPrefix: true,
			},
		},
	})
	assert.NilError(t, err)

	assert.DeepEqual(t, values, &apiv1.GetValuesResponse{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2"),
			},
		},
	})
}

func Test_Get(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "Test_GetPrefix")
	assert.NilError(t, err)

	store, err := datastore.NewKVStore(tmpDir)
	assert.NilError(t, err)

	err = store.Set(&apiv1.SetValuesRequest{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2 longer"),
			},
			&apiv1.KeyValue{
				Key:   "test/2/stuff",
				Value: []byte("bad value"),
			},
		},
	})
	assert.NilError(t, err)

	values, err := store.Get(&apiv1.GetValuesRequest{
		Requests: []*apiv1.GetValuesRequest_GetValue{
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff2",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/2/stuff",
			},
		},
	})
	assert.NilError(t, err)

	assert.DeepEqual(t, values, &apiv1.GetValuesResponse{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2 longer"),
			},
			&apiv1.KeyValue{
				Key:   "test/2/stuff",
				Value: []byte("bad value"),
			},
		},
	})
}

func Test_TTL(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "Test_GetPrefix")
	assert.NilError(t, err)

	store, err := datastore.NewKVStore(tmpDir)
	assert.NilError(t, err)

	ttl := 2 * time.Second
	err = store.Set(&apiv1.SetValuesRequest{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2 longer"),
			},
			&apiv1.KeyValue{
				Key:   "test/2/stuff",
				Value: []byte("bad value"),
			},
		},
		TtlDuration: ptypes.DurationProto(ttl),
	})
	assert.NilError(t, err)

	values, err := store.Get(&apiv1.GetValuesRequest{
		Requests: []*apiv1.GetValuesRequest_GetValue{
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff2",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/2/stuff",
			},
		},
	})
	assert.NilError(t, err)

	assert.DeepEqual(t, values, &apiv1.GetValuesResponse{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "test/1/stuff",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2 longer"),
			},
			&apiv1.KeyValue{
				Key:   "test/2/stuff",
				Value: []byte("bad value"),
			},
		},
	})

	time.Sleep(ttl)
	values, err = store.Get(&apiv1.GetValuesRequest{
		Requests: []*apiv1.GetValuesRequest_GetValue{
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/1/stuff2",
			},
			&apiv1.GetValuesRequest_GetValue{
				Key: "test/2/stuff",
			},
		},
	})
	assert.NilError(t, err)
	assert.Equal(t, len(values.Messages), 0)
}

func Test_Subscribe(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "Test_GetPrefix")
	assert.NilError(t, err)

	store, err := datastore.NewKVStore(tmpDir)
	assert.NilError(t, err)

	values := []*apiv1.KeyValue{}
	mtx := sync.Mutex{}

	err = store.Set(&apiv1.SetValuesRequest{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "subscribe/5/serial1",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "test/1/stuff2",
				Value: []byte("value 2"),
			},
		},
	})
	assert.NilError(t, err)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		store.Subscribe(ctx, &apiv1.SubscribeRequest{
			Prefixes: []string{"subscribe/5"},
		}, func(msg *apiv1.SubscribeResponse) error {
			mtx.Lock()
			defer mtx.Unlock()
			fmt.Println("here")
			values = append(values, msg.Messages...)
			return nil
		})
	}()

	time.Sleep(10 * time.Millisecond)

	err = store.Set(&apiv1.SetValuesRequest{
		Messages: []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "subscribe/5/serial1",
				Value: []byte("value 1 1"),
			},
		},
	})
	assert.NilError(t, err)

	now := time.Now()
	for {
		if time.Now().Sub(now) > 30*time.Second {
			t.Fatal("timed out waiting for results")
		}

		count := 0
		mtx.Lock()
		count = len(values)
		mtx.Unlock()

		if count != 2 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		assert.DeepEqual(t, values, []*apiv1.KeyValue{
			&apiv1.KeyValue{
				Key:   "subscribe/5/serial1",
				Value: []byte("value 1"),
			},
			&apiv1.KeyValue{
				Key:   "subscribe/5/serial1",
				Value: []byte("value 1 1"),
			},
		})
		break
	}
}
