package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/service"
)

func main() {
	srvc := service.NewService()
	go srvc.ProcessDeletions()
	defer srvc.Helper.Close()

	myHandler := handler.NewHandler(srvc)
	err := http.ListenAndServe(srvc.Config.ServerURL, myHandler.Router())
	if err != nil {
		panic(err)
	}
}
