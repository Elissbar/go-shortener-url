package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
)

func main() {
	var urls map[string]string

	// Файл используется для хранения коротких ссылок
	err := handler.GetAllLinks(&urls)
	if err != nil {
		panic("Error: " + err.Error())
	}

	// Структуру используем чтобы не открывать файл при каждом POST/GET запросе
	myHandler := handler.MyHandler{Urls: urls}

	mux := http.NewServeMux()
	mux.HandleFunc("/", myHandler.CreateShortUrl)
	mux.HandleFunc("/{id}", myHandler.GetShortUrl)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
