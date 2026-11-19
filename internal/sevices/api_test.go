package services_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	services "github.com/syncromatics/kvetch/internal/sevices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gotest.tools/assert"
)

type fakeDatastore struct {
	subscribe func(ctx context.Context, cb func(*apiv1.SubscribeResponse) error) error
}

func (d *fakeDatastore) Get(*apiv1.GetValuesRequest) (*apiv1.GetValuesResponse, error) {
	return nil, nil
}

func (d *fakeDatastore) Set(*apiv1.SetValuesRequest) error {
	return nil
}

func (d *fakeDatastore) Subscribe(ctx context.Context, _ *apiv1.SubscribeRequest, cb func(*apiv1.SubscribeResponse) error) error {
	return d.subscribe(ctx, cb)
}

type fakeSubscribeServer struct {
	grpc.ServerStream
	ctx     context.Context
	sendErr error
	onSend  func()
}

func (s *fakeSubscribeServer) Context() context.Context {
	return s.ctx
}

func (s *fakeSubscribeServer) Send(*apiv1.SubscribeResponse) error {
	if s.onSend != nil {
		s.onSend()
	}
	return s.sendErr
}

func Test_Subscribe_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := services.NewAPIService(&fakeDatastore{
		subscribe: func(ctx context.Context, cb func(*apiv1.SubscribeResponse) error) error {
			return errors.Wrap(ctx.Err(), "failed to subscribe")
		},
	})

	err := service.Subscribe(&apiv1.SubscribeRequest{}, &fakeSubscribeServer{ctx: ctx})
	assert.Equal(t, status.Code(err), codes.Canceled)
}

func Test_Subscribe_SendFailsAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	service := services.NewAPIService(&fakeDatastore{
		subscribe: func(ctx context.Context, cb func(*apiv1.SubscribeResponse) error) error {
			return errors.Wrap(cb(&apiv1.SubscribeResponse{}), "failed callback")
		},
	})

	err := service.Subscribe(&apiv1.SubscribeRequest{}, &fakeSubscribeServer{
		ctx:     ctx,
		onSend:  cancel,
		sendErr: status.Error(codes.Canceled, "context canceled"),
	})
	assert.Equal(t, status.Code(err), codes.Canceled)
}

func Test_Subscribe_DeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	<-ctx.Done()

	service := services.NewAPIService(&fakeDatastore{
		subscribe: func(ctx context.Context, cb func(*apiv1.SubscribeResponse) error) error {
			return errors.Wrap(ctx.Err(), "failed to subscribe")
		},
	})

	err := service.Subscribe(&apiv1.SubscribeRequest{}, &fakeSubscribeServer{ctx: ctx})
	assert.Equal(t, status.Code(err), codes.DeadlineExceeded)
}

func Test_Subscribe_DatastoreFailureIsReported(t *testing.T) {
	service := services.NewAPIService(&fakeDatastore{
		subscribe: func(ctx context.Context, cb func(*apiv1.SubscribeResponse) error) error {
			return errors.New("badger exploded")
		},
	})

	err := service.Subscribe(&apiv1.SubscribeRequest{}, &fakeSubscribeServer{ctx: context.Background()})
	assert.Equal(t, status.Code(err), codes.Unknown)
	assert.ErrorContains(t, err, "badger exploded")
}
