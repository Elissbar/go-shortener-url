package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/handler"
)

func main() {
	parseFlags()
	config := config.New(serverUrl, baseUrl)

	urls := make(map[string]string)
	myHandler := handler.MyHandler{Urls: urls, Config: config}

	router := myHandler.Router()

	err := http.ListenAndServe(serverUrl, router)
	if err != nil {
		panic(err)
	}
}
