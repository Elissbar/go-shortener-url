package main

import (
	"flag"
)

var serverUrl, baseUrl string

func parseFlags() {
	flag.StringVar(&serverUrl, "a", "localhost:8080", "<host>:<port>")
	flag.StringVar(&baseUrl, "b", "http://localhost:8080/", "Base URL for the API. Example: http://localhost:8080/")
	flag.Parse()
}
