package grpc

import (
	"google.golang.org/grpc"
)

// WithInsecure returns a DialOption which disables transport security for this
// ClientConn. Note that transport security is required unless WithInsecure is
// set.
func WithInsecure() grpc.DialOption {
	return grpc.WithInsecure()
}
