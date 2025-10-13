package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
)

func main() {
	urls := make(map[string]string)
	myHandler := handler.MyHandler{Urls: urls}

	router := myHandler.Router()

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}
