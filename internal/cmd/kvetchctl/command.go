package kvetchctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var (
	endpoint     string
	isPrefix     bool = false
	outputFormat string
	ttl          time.Duration
	valueType    string
	verbose      bool = false

	log     *zap.SugaredLogger
	client  apiv1.APIClient
	rootCmd = &cobra.Command{
		Use:               "kvetchctl",
		Short:             "Command line interface for interacting with Kvetch",
		PersistentPreRunE: setupLogger,
	}
)

type jsonOutput struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func init() {
	setupLogger(nil, nil)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "e", "", "Kvetch instance to connect to")
	rootCmd.MarkPersistentFlagRequired("endpoint")

	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVarP(&isPrefix, "prefix", "p", false, "Treat the given keys as prefixes")
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "simple", "Set the output format (simple, json)")
	getCmd.Flags().StringVarP(&valueType, "value-type", "t", "string", "Set the type of value in the output (string, bytes, json)")

	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().StringVarP(&outputFormat, "output", "o", "simple", "Set the output format (simple, json)")
	watchCmd.Flags().StringVarP(&valueType, "value-type", "t", "string", "Set the type of value in the output (string, bytes, json)")

	rootCmd.AddCommand(setCmd)
	setCmd.Flags().DurationVar(&ttl, "ttl", 0, "Set the time-to-live for each key (optional)")
	setCmd.Flags().StringVarP(&outputFormat, "output", "o", "simple", "Set the output format (simple, json)")
	setCmd.Flags().StringVarP(&valueType, "value-type", "t", "string", "Set the type of value in the input (string, bytes, json)")
}

// Execute executes the command line interface
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Infow("exited with error", "err", err)
		os.Exit(1)
	}
}

func setupLogger(*cobra.Command, []string) error {
	config := zap.NewDevelopmentConfig()
	config.Encoding = "console"
	if verbose {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	log = logger.Sugar()
	return nil
}

func setupClient(*cobra.Command, []string) error {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to connect to endpoint")
	}

	client = apiv1.NewAPIClient(conn)
	return nil
}

func writeOutput(ctx context.Context, messages []*apiv1.KeyValue, output *os.File) error {
	for _, message := range messages {
		select {
		case <-ctx.Done():
			return io.ErrClosedPipe
		default:
		}

		switch outputFormat {
		case "simple":
			output.WriteString(message.Key)
			output.WriteString(": ")
			output.Write(message.Value)
			output.WriteString("\n")
			break
		case "json":
			m := map[string]interface{}{}
			switch valueType {
			case "string":
				m[message.Key] = string(message.Value)
				break
			case "bytes":
				m[message.Key] = message.Value
				break
			case "json":
				var j interface{}
				err := json.Unmarshal(message.Value, &j)
				m[message.Key] = j
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to unmarshal value for key %s", message.Key))
				}
			default:
				return errors.New("not implemented")
			}

			bytes, err := json.Marshal(m)
			if err != nil {
				return errors.Wrap(err, "failed to marshal value")
			}

			output.Write(bytes)
			output.WriteString("\n")
			break
		default:
			return errors.New("not implemented")
		}
	}
	return nil
}
