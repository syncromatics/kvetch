package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	kvstore "github.com/syncromatics/kvetch/internal/datastore"
)

type settings struct {
	Port                      int
	PrometheusPort            int
	Datastore                 string
	GarbageCollectionInterval time.Duration
	KVStoreOptions            *kvstore.KVStoreOptions
}

func getKVStoreOptions() (*kvstore.KVStoreOptions, error) {
	allErrors := []string{}
	kvStoreOptions := &kvstore.KVStoreOptions{}
	inMemoryString, ok := os.LookupEnv("IN_MEMORY")
	if ok {
		inMemory, err := strconv.ParseBool(inMemoryString)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("IN_MEMORY is not a valid bool '%s'", inMemoryString))
		} else {
			kvStoreOptions.InMemory = &wrappers.BoolValue{Value: inMemory}
		}
	} else {
		kvStoreOptions.InMemory = &wrappers.BoolValue{Value: false}
	}
	enableTruncateString, ok := os.LookupEnv("ENABLE_TRUNCATE")
	if ok {
		enableTruncate, err := strconv.ParseBool(enableTruncateString)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("ENABLE_TRUNCATE is not a valid bool '%s'", enableTruncateString))
		} else {
			kvStoreOptions.EnableTruncate = &wrappers.BoolValue{Value: enableTruncate}
		}
	}
	maxTableSizeString, ok := os.LookupEnv("MAX_TABLE_SIZE")
	if ok {
		maxTableSize, err := strconv.ParseInt(maxTableSizeString, 10, 64)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("MAX_TABLE_SIZE is not a valid int64 '%s'", maxTableSizeString))
		} else {
			kvStoreOptions.MaxTableSize = &wrappers.Int64Value{Value: maxTableSize}
		}
	}
	levelOneSizeString, ok := os.LookupEnv("LEVEL_ONE_SIZE")
	if ok {
		levelOneSize, err := strconv.ParseInt(levelOneSizeString, 10, 64)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("LEVEL_ONE_SIZE is not a valid int64 '%s'", levelOneSizeString))
		} else {
			kvStoreOptions.LevelOneSize = &wrappers.Int64Value{Value: levelOneSize}
		}
	}
	levelSizeMultiplierString, ok := os.LookupEnv("LEVEL_SIZE_MULTIPLIER")
	if ok {
		levelSizeMultiplier, err := strconv.ParseInt(levelSizeMultiplierString, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("LEVEL_SIZE_MULTIPLIER is not a valid int32 '%s'", levelSizeMultiplierString))
		} else {
			kvStoreOptions.LevelSizeMultiplier = &wrappers.Int32Value{Value: int32(levelSizeMultiplier)}
		}
	}
	numberOfLevelZeroTablesString, ok := os.LookupEnv("NUMBER_OF_LEVEL_ZERO_TABLES")
	if ok {
		numberOfLevelZeroTables, err := strconv.ParseInt(numberOfLevelZeroTablesString, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("NUMBER_OF_LEVEL_ZERO_TABLES is not a valid int32 '%s'", numberOfLevelZeroTablesString))
		} else {
			kvStoreOptions.NumberOfLevelZeroTables = &wrappers.Int32Value{Value: int32(numberOfLevelZeroTables)}
		}
	}
	numberOfLevelZeroTablesUntilForceCompactionString, ok := os.LookupEnv("NUMBER_OF_ZERO_LEVEL_TABLES_UNTIL_FORCE_COMPACTION")
	if ok {
		numberOfLevelZeroTablesUntilForceCompactaion, err := strconv.ParseInt(numberOfLevelZeroTablesUntilForceCompactionString, 10, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("NUMBER_OF_ZERO_LEVEL_TABLES_UNTIL_FORCE_COMPACTION is not a valid int32 '%s'", numberOfLevelZeroTablesUntilForceCompactionString))
		} else {
			kvStoreOptions.NumberOfLevelZeroTablesUntilForceCompaction = &wrappers.Int32Value{Value: int32(numberOfLevelZeroTablesUntilForceCompactaion)}
		}
	}
	garbageCollectionDiscardRatioString, ok := os.LookupEnv("GARBAGE_COLLECTION_DISCARD_RATIO")
	if ok {
		garbageCollectionDiscardRatio, err := strconv.ParseFloat(garbageCollectionDiscardRatioString, 32)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("GARBAGE_COLLECTION_DISCARD_RATIO is not a valid float32 '%s'", garbageCollectionDiscardRatioString))
		} else {
			kvStoreOptions.GarbageCollectionDiscardRatio = &wrappers.FloatValue{Value: float32(garbageCollectionDiscardRatio)}
		}
	}

	if len(allErrors) > 0 {
		return nil, fmt.Errorf("Failed configuring KVStore: %s", strings.Join(allErrors, ", "))
	}

	return kvStoreOptions, nil
}

func getSettingsFromEnv() (*settings, error) {
	allErrors := []string{}
	var err error

	portInt := 0
	port, ok := os.LookupEnv("PORT")
	if !ok {
		portInt = 7777
	} else {
		portInt, err = strconv.Atoi(port)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("failed to convert %s to int", port))
		}
	}

	prometheusPortInt := 0
	port, ok = os.LookupEnv("PROMETHEUS_PORT")
	if !ok {
		prometheusPortInt = 80
	} else {
		prometheusPortInt, err = strconv.Atoi(port)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("failed to convert %s to int", port))
		}
	}

	kvStoreOptions, err := getKVStoreOptions()
	if err != nil {
		allErrors = append(allErrors, err.Error())
	}

	datastore, ok := os.LookupEnv("DATASTORE")
	if !ok && !kvStoreOptions.InMemory.Value {
		allErrors = append(allErrors, "DATASTORE")
	}

	duration := 5 * time.Minute
	collection, ok := os.LookupEnv("GARBAGE_COLLECTION_INTERVAL")
	if ok {
		duration, err = time.ParseDuration(collection)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("GARBAGE_COLLECTION_INTERVAL is not a valid time.Duration '%s'", collection))
		}
	}

	if len(allErrors) > 0 {
		return nil, fmt.Errorf("Missing required environment variables: %s", strings.Join(allErrors, ", "))
	}

	return &settings{
		Port:                      portInt,
		PrometheusPort:            prometheusPortInt,
		Datastore:                 datastore,
		GarbageCollectionInterval: duration,
		KVStoreOptions:            kvStoreOptions,
	}, nil
}
