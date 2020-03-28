package services

import (
	"context"

	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"

	"github.com/pkg/errors"
)

// Datastore is the key value datastore.
type Datastore interface {
	Get(request *apiv1.GetValuesRequest) (*apiv1.GetValuesResponse, error)
	Set(request *apiv1.SetValuesRequest) error
	Subscribe(ctx context.Context, subscription *apiv1.SubscribeRequest, cb func(*apiv1.SubscribeResponse) error) error
}

var _ apiv1.APIServer = &APIService{}

// APIService is the grpc service on top of the datastore
type APIService struct {
	datastore Datastore
}

// NewAPIService creates a new api service
func NewAPIService(datastore Datastore) *APIService {
	return &APIService{datastore}
}

// GetValues gets a list of key values
func (s *APIService) GetValues(ctx context.Context, request *apiv1.GetValuesRequest) (*apiv1.GetValuesResponse, error) {
	r, err := s.datastore.Get(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get from datastore")
	}

	return r, nil
}

// SetValues sets a list of key values
func (s *APIService) SetValues(ctx context.Context, request *apiv1.SetValuesRequest) (*apiv1.SetValuesResponse, error) {
	err := s.datastore.Set(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set in datastore")
	}

	return &apiv1.SetValuesResponse{}, nil
}

// Subscribe subscribes to a list of prefixes
func (s *APIService) Subscribe(request *apiv1.SubscribeRequest, stream apiv1.API_SubscribeServer) error {
	err := s.datastore.Subscribe(stream.Context(), request, stream.Send)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}
	return nil
}
