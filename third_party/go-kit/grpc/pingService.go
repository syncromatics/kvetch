package grpc

import (
	pingv1 "github.com/syncromatics/go-kit/internal/protos/gokit/ping/v1"

	"golang.org/x/net/context"
)

type pingService struct{}

func (*pingService) Ping(ctx context.Context, m *pingv1.PingRequest) (*pingv1.PingResponse, error) {
	return &pingv1.PingResponse{}, nil
}
