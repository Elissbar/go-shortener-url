package main

import (
	"flag"
	"os"

	"github.com/Elissbar/go-shortener-url/internal/config"
)

func parseFlags() *config.Config {
	var serverUrl, baseUrl string

	flag.StringVar(&serverUrl, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseUrl, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.Parse()

	if osEnvServerAddr := os.Getenv("SERVER_ADDRESS"); osEnvServerAddr != "" {
		serverUrl = osEnvServerAddr
	}
	if osEnvBaseUrl := os.Getenv("BASE_URL"); osEnvBaseUrl != "" {
		baseUrl = osEnvBaseUrl
	}

	return config.New(serverUrl, baseUrl)
}
