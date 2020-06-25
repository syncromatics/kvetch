package kvetchctl

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syncromatics/go-kit/cmd"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	getCmd = &cobra.Command{
		Use:     "get [flags] [keys/prefix]",
		Short:   "Get values by key or prefix",
		Args:    cobra.MinimumNArgs(1),
		PreRunE: setupClient,
		RunE: func(_ *cobra.Command, args []string) error {
			isPrefix := viper.GetBool("prefix")
			group := cmd.NewProcessGroup(context.Background())
			for _, k := range args {
				key := k
				group.Go(func() error {
					logger := log.With(
						"key", key,
						"isPrefix", isPrefix,
					)
					logger.Info("getting key")
					response, err := client.GetValues(group.Context(), &apiv1.GetValuesRequest{
						Requests: []*apiv1.GetValuesRequest_GetValue{
							{
								Key:      key,
								IsPrefix: isPrefix,
							},
						},
					})
					s, ok := status.FromError(err)
					if ok && s.Code() == codes.Canceled {
						return nil
					}
					if err != nil {
						logger.Error(err, "failed to get key")
						return errors.Wrap(err, fmt.Sprintf("failed to get value(s) by key %s", key))
					}
					return writeOutput(group.Context(), response.Messages, os.Stdout)
				})
			}

			return group.Wait()
		},
	}
)

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("prefix", "p", false, "Treat the given keys as prefixes")
	bindCommonFlags(getCmd)
}
