package kvetchctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var (
	log     *zap.SugaredLogger
	client  apiv1.APIClient
	rootCmd = &cobra.Command{
		Use:               "kvetchctl",
		Short:             "Command line interface for interacting with Kvetch",
		PersistentPreRunE: setupLogger,
	}
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("kvetchctl")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	setupLogger(rootCmd, nil)

	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("prefix", "p", false, "Treat the given keys as prefixes")
	bindCommonFlags(getCmd)

	rootCmd.AddCommand(watchCmd)
	bindCommonFlags(watchCmd)

	rootCmd.AddCommand(setCmd)
	setCmd.Flags().Duration("ttl", 0, "Set the time-to-live for each key (optional)")
	bindCommonFlags(setCmd)
}

func bindCommonFlags(command *cobra.Command) {
	command.Flags().StringP("endpoint", "e", "", "Kvetch instance to connect to (required)")
	command.Flags().StringP("output", "o", "simple", "Set the output format (simple, json)")
	command.Flags().StringP("value-type", "t", "string", "Set the type of value in the output (string, bytes, json)")
}

// Execute executes the command line interface
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Infow("exited with error", "err", err)
		os.Exit(1)
	}
}

func setupLogger(command *cobra.Command, _ []string) error {
	err := viper.BindPFlags(command.PersistentFlags())
	if err != nil {
		return err
	}

	verbose := viper.GetBool("verbose")
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

func setupClient(command *cobra.Command, args []string) error {
	err := viper.BindPFlags(command.Flags())
	if err != nil {
		return err
	}

	endpoint := viper.GetString("endpoint")
	if endpoint == "" {
		return errors.New(`required flag "endpoint" not set`)
	}
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to connect to endpoint")
	}

	client = apiv1.NewAPIClient(conn)
	return nil
}

func writeOutput(ctx context.Context, messages []*apiv1.KeyValue, output *os.File) error {
	outputFormat := viper.GetString("output")
	valueType := viper.GetString("value-type")
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
