package grpc

import (
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	streamingCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "grpc_server_active_streams",
		Help: "The total number of streams connected to the service",
	}, []string{
		"grpc_type",
		"grpc_service",
		"grpc_method",
	})
)

type timingWrapper struct {
	stream   grpc.ServerStream
	ctx      context.Context
	cancel   func()
	timeout  time.Duration
	received chan struct{}
	labels   prometheus.Labels
}

func newTimingWrapper(stream grpc.ServerStream, timeout time.Duration, serviceName, serviceMethod string) *timingWrapper {
	ctx, cancel := context.WithCancel(stream.Context())
	received := make(chan struct{})

	labels := prometheus.Labels{
		"grpc_type":    string(grpc_prometheus.BidiStream),
		"grpc_service": serviceName,
		"grpc_method":  serviceMethod,
	}

	return &timingWrapper{
		stream:   stream,
		ctx:      ctx,
		cancel:   cancel,
		timeout:  timeout,
		received: received,
		labels:   labels,
	}
}

func (w *timingWrapper) Watch() {
	streamingCount.With(w.labels).Inc()
	defer streamingCount.With(w.labels).Dec()

	for {
		timeout := time.NewTimer(w.timeout)
		select {
		case <-w.ctx.Done():
			return
		case <-timeout.C:
			w.cancel()
			return
		case <-w.received:
			continue
		}
	}
}

func (w *timingWrapper) Context() context.Context {
	return w.ctx
}

func (w *timingWrapper) RecvMsg(m interface{}) error {
	err := w.stream.RecvMsg(m)
	if err != nil {
		return err
	}

	select {
	case w.received <- struct{}{}:
	case <-w.ctx.Done():
	}

	return nil
}

func (w *timingWrapper) SendMsg(m interface{}) error {
	return w.stream.SendMsg(m)
}

func (w *timingWrapper) SendHeader(metadata metadata.MD) error {
	return w.stream.SendHeader(metadata)
}

func (w *timingWrapper) SetHeader(metadata metadata.MD) error {
	return w.stream.SetHeader(metadata)
}

func (w *timingWrapper) SetTrailer(metadata metadata.MD) {
	w.stream.SetTrailer(metadata)
}

func streamTimingInterceptor(serviceName string, timeout time.Duration) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !info.IsClientStream {
			return handler(srv, stream)
		}

		if !info.IsServerStream {
			return handler(srv, stream)
		}

		//only handle bidirectional streams

		wrapper := newTimingWrapper(stream, timeout, serviceName, info.FullMethod)

		go wrapper.Watch()

		return handler(srv, wrapper)
	}
}
