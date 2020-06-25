package kvetchctl

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/syncromatics/go-kit/cmd"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	watchCmd = &cobra.Command{
		Use:     "watch [flags] [key prefixes]",
		Short:   "Watch values by prefix",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: setupClient,
		RunE: func(_ *cobra.Command, prefixes []string) error {
			group := cmd.NewProcessGroup(context.Background())
			group.Go(func() error {
				logger := log.With(
					"prefixes", prefixes,
				)
				logger.Info("watching prefixes")
				stream, err := client.Subscribe(group.Context(), &apiv1.SubscribeRequest{
					Prefixes: prefixes,
				})
				if err != nil {
					logger.Error(err, "failed to watch prefixes")
					return errors.Wrap(err, fmt.Sprintf("failed to get value(s) by key %s", prefixes))
				}
				for {
					response, err := stream.Recv()
					s, ok := status.FromError(err)
					if ok && s.Code() == codes.Canceled {
						return nil
					}
					if err != nil {
						return errors.Wrap(err, "failed to read response")
					}

					err = writeOutput(stream.Context(), response.Messages, os.Stdout)
					if err != nil {
						return err
					}
				}
			})

			return group.Wait()
		},
	}
)

func init() {
	RootCmd.AddCommand(watchCmd)
	bindCommonFlags(watchCmd)
}
