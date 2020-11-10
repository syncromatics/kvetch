package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/syncromatics/kvetch/internal/datastore"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	services "github.com/syncromatics/kvetch/internal/sevices"

	"github.com/syncromatics/go-kit/grpc"
	"golang.org/x/sync/errgroup"
)

func main() {
	settings, err := getSettingsFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	kvstore, err := datastore.NewKVStore(settings.Datastore, settings.KVStoreOptions)
	if err != nil {
		log.Fatal(err)
	}

	service := services.NewAPIService(kvstore)

	server := grpc.CreateServer(&grpc.Settings{
		ServerName: "kvetch",
	})

	apiv1.RegisterAPIServer(server, service)

	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)

	group.Go(grpc.HostServer(ctx, server, settings.Port))
	group.Go(grpc.HostMetrics(ctx, settings.PrometheusPort))

	if !settings.KVStoreOptions.InMemory.GetValue() {
		garbageCollector := services.NewGarbageCollectorService(kvstore, settings.GarbageCollectionInterval)
		group.Go(garbageCollector.Run(ctx))
	}

	eventChan := make(chan os.Signal)
	signal.Notify(eventChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("kvetch started...")

	select {
	case <-eventChan:
	case <-ctx.Done():
	}

	fmt.Println("kvetch stopping...")

	cancel()

	if err := group.Wait(); err != nil {
		log.Fatal(err)
	}
}
