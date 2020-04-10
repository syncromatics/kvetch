package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type settings struct {
	Port                      int
	PrometheusPort            int
	Datastore                 string
	GarbageCollectionInterval time.Duration
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

	datastore, ok := os.LookupEnv("DATASTORE")
	if !ok {
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
	}, nil
}
