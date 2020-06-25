package kvetchctl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/syncromatics/go-kit/cmd"
	apiv1 "github.com/syncromatics/kvetch/internal/protos/kvetch/api/v1"
)

var (
	setCmd = &cobra.Command{
		Use: `set [flags] [key] [value]
-or- set [flags] [key]
-or- set [flags]`,
		Short: "Set values by key",
		Long: `Sets values for keys

If both a key and value are specified as arguments, the key will be set with the given value.
If only a key is specified as an argument, the key will be set with the value read from STDIN.
If neither a key nor value is specified as an argument, the keys and values will be read from JSON objects from STDIN.`,
		Args:    cobra.RangeArgs(0, 2),
		PreRunE: setupClient,
		RunE: func(_ *cobra.Command, args []string) error {
			group := cmd.NewProcessGroup(context.Background())
			group.Go(func() error {
				ttl := viper.GetDuration("ttl")
				var ttlDuration *duration.Duration
				if ttl != 0 {
					ttlDuration = ptypes.DurationProto(ttl)
				}

				switch len(args) {
				case 0:
					return setWithZeroArgs(group.Context(), ttlDuration)
				case 1:
					return setWithOneArg(group.Context(), ttlDuration, args[0])
				case 2:
					return setWithTwoArgs(group.Context(), ttlDuration, args[0], args[1])
				default:
					return errors.New("not implemented")
				}
			})

			return group.Wait()
		},
	}
)

func init() {
	RootCmd.AddCommand(setCmd)
	setCmd.Flags().Duration("ttl", 0, "Set the time-to-live for each key (optional)")
	bindCommonFlags(setCmd)
}

func setWithZeroArgs(ctx context.Context, ttlDuration *duration.Duration) error {
	valueType := viper.GetString("value-type")
	decoder := json.NewDecoder(os.Stdin)
	for count := 0; ; count++ {
		select {
		case <-ctx.Done():
			return io.ErrClosedPipe
		default:
			if !decoder.More() {
				return nil
			}
		}

		var valueObjects map[string]interface{}
		err := decoder.Decode(&valueObjects)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "failed to decode stdin as JSON")
		}

		messages := []*apiv1.KeyValue{}
		for key, valueObject := range valueObjects {
			var value []byte
			switch valueType {
			case "json":
				value, err = json.Marshal(valueObject)
				if err != nil {
					return errors.Wrap(err, fmt.Sprintf("failed to marshal value for key %s", key))
				}
				break
			case "string":
				switch valueString := valueObject.(type) {
				case string:
					value = []byte(valueString)
				default:
					return fmt.Errorf("value for key %s is not a string", key)
				}
			case "bytes":
				switch valueString := valueObject.(type) {
				case string:
					value, err = base64.StdEncoding.DecodeString(valueString)
				default:
					return fmt.Errorf("value for key %s is not a base64-encoded string", key)
				}
			default:
				return errors.New("not implemented")
			}
			messages = append(messages, &apiv1.KeyValue{
				Key:   key,
				Value: value,
			})
		}
		err = setValues(ctx, ttlDuration, messages)
		if err != nil {
			return err
		}
	}
}

func setWithOneArg(ctx context.Context, ttlDuration *duration.Duration, key string) error {
	valueType := viper.GetString("value-type")
	value, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrap(err, "failed to read value from stdin")
	}
	switch valueType {
	case "string":
		// NOOP
		break
	case "bytes":
		value, err = base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			return fmt.Errorf("value for key %s is not a base64-encoded string", key)
		}
	default:
		return fmt.Errorf("value type %s is not supported in this case", valueType)
	}
	return setValues(ctx, ttlDuration, []*apiv1.KeyValue{
		{
			Key:   key,
			Value: value,
		},
	})
}

func setWithTwoArgs(ctx context.Context, ttlDuration *duration.Duration, key, valueString string) error {
	valueType := viper.GetString("value-type")
	var err error
	var value []byte
	switch valueType {
	case "string":
		value = []byte(valueString)
		break
	case "bytes":
		value, err = base64.StdEncoding.DecodeString(valueString)
		if err != nil {
			return fmt.Errorf("value for key %s is not a base64-encoded string", key)
		}
	default:
		return fmt.Errorf("value type %s is not supported in this case", valueType)
	}
	return setValues(ctx, ttlDuration, []*apiv1.KeyValue{
		{
			Key:   key,
			Value: value,
		},
	})
}

func setValues(ctx context.Context, ttlDuration *duration.Duration, messages []*apiv1.KeyValue) error {
	_, err := client.SetValues(ctx, &apiv1.SetValuesRequest{
		TtlDuration: ttlDuration,
		Messages:    messages,
	})
	if err != nil {
		return errors.Wrap(err, "failed to set values")
	}
	return writeOutput(ctx, messages, os.Stdout)
}
