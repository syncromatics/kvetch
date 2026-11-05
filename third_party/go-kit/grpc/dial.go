package grpc

import (
	"google.golang.org/grpc"
)

// Dial creates a client connection to the given target
func Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(target, opts...)
}
