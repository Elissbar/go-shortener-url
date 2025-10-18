package main

import (
	"flag"

	"github.com/Elissbar/go-shortener-url/internal/config"
)

func parseFlags() *config.Config {
	var serverUrl, baseUrl string

	flag.StringVar(&serverUrl, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseUrl, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.Parse()

	return config.New(serverUrl, baseUrl)
}
