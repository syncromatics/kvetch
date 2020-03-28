package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type settings struct {
	Port           int
	PrometheusPort int
	Datastore      string
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
		portInt, err = strconv.Atoi(port)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("failed to convert %s to int", port))
		}
	}

	datastore, ok := os.LookupEnv("DATASTORE")
	if !ok {
		allErrors = append(allErrors, "DATASTORE")
	}

	if len(allErrors) > 0 {
		return nil, fmt.Errorf("Missing required environment variables: %s", strings.Join(allErrors, ", "))
	}

	return &settings{
		Port:           portInt,
		PrometheusPort: prometheusPortInt,
		Datastore:      datastore,
	}, nil
}
