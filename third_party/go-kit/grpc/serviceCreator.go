package grpc

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	pingv1 "github.com/syncromatics/go-kit/internal/protos/gokit/ping/v1"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaegerClient "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/thrift-gen/sampling"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Settings are the settings for the grpc service
type Settings struct {
	JaegerAgentHost            string
	DefaultUDPSpanServerPort   string
	ServerName                 string
	LogLevel                   string
	BiDirectionalStreamTimeout time.Duration
	Sampler                    jaegerClient.Sampler
}

// CreateServer will create a grpc server with tracing, prometheus stats, and logging
func CreateServer(s *Settings) *grpc.Server {
	logConfig := zap.NewProductionConfig()
	logLevel := zap.WarnLevel
	if s.LogLevel != "" {
		if err := logLevel.UnmarshalText([]byte(s.LogLevel)); err != nil {
			panic(fmt.Errorf("invalid gRPC log level %q: %w", s.LogLevel, err))
		}
	}
	logConfig.Level = zap.NewAtomicLevelAt(logLevel)
	logger, _ := logConfig.Build()

	if s.BiDirectionalStreamTimeout == 0 {
		s.BiDirectionalStreamTimeout = 1 * time.Minute
	}

	grpc_zap.ReplaceGrpcLogger(logger)

	transport, err := jaegerClient.NewUDPTransport(fmt.Sprintf("%s:%d", s.JaegerAgentHost, jaegerClient.DefaultUDPSpanServerPort), 60000)
	if err != nil {
		panic(err)
	}

	sampler := s.Sampler
	if sampler == nil {
		sampler = jaegerClient.NewPerOperationSampler(jaegerClient.PerOperationSamplerParams{
			Strategies: &sampling.PerOperationSamplingStrategies{
				DefaultSamplingProbability:       0.1,
				DefaultLowerBoundTracesPerSecond: 1.0,
			},
		})
	}

	tracer, _ := jaegerClient.NewTracer(s.ServerName,
		sampler,
		jaegerClient.NewRemoteReporter(transport))

	opentracing.SetGlobalTracer(tracer)
	grpc_prometheus.EnableHandlingTimeHistogram()

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
			streamTimingInterceptor(s.ServerName, s.BiDirectionalStreamTimeout),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              10 * time.Second, // wait time before ping if no activity
			Timeout:           20 * time.Second, // ping timeout
			MaxConnectionIdle: 2 * time.Minute,  // Max time a connection can be idle, includes pings

		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: 10 * time.Second, // min time a client should wait before sending a ping
		}),
		grpc.MaxConcurrentStreams(math.MaxUint32))

	pingv1.RegisterPingAPIServer(server, &pingService{})

	return server
}

// HostServer will host the grpc server and gracefully stop if the context is completed
func HostServer(ctx context.Context, server *grpc.Server, port int) func() error {
	cancel := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			go func() {
				time.Sleep(10 * time.Second)
				server.Stop()
			}()
			server.GracefulStop()
			return
		case <-cancel:
			return
		}
	}()

	return func() error {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			close(cancel)
			return errors.Wrap(err, "failed to listen")
		}

		reflection.Register(server)

		if err := server.Serve(lis); err != nil {
			close(cancel)
			return errors.Wrap(err, "failed to serve")
		}

		return nil
	}
}

// HostMetrics will host the prometheus metrics
func HostMetrics(ctx context.Context, port int) func() error {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.Handle("/metrics", promhttp.Handler())

	cancel := make(chan error)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			cancel <- errors.Wrap(err, "failed to serve http")
		}
	}()

	return func() error {
		select {
		case <-ctx.Done():
			// do nothing for now
			return nil
		case msg := <-cancel:
			return msg
		}
	}
}

// WaitTillServiceIsAvailable uses the ping service to wait till a grpc server is available
func WaitTillServiceIsAvailable(host string, port int, duration time.Duration) error {
	start := time.Now()
	for time.Now().Sub(start) < duration {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
		if err != nil {
			continue
		}

		client := pingv1.NewPingAPIClient(conn)
		_, err = client.Ping(context.Background(), &pingv1.PingRequest{})
		conn.Close()

		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed waiting for grpc server to start")
}
