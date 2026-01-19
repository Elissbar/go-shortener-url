package main

import (
	"net/http"

	"github.com/Elissbar/go-shortener-url/internal/handler"
	"github.com/Elissbar/go-shortener-url/internal/service"
	// _ "net/http/pprof"
)

func main() {
	srvc := service.NewService()
	go srvc.ProcessDeletions()
	defer srvc.Helper.Close()

	// для запуска pprof на отдельном порту
	// go func() {
    //     http.ListenAndServe("localhost:6060", nil)
    // }()

	myHandler := handler.NewHandler(srvc)
	err := http.ListenAndServe(srvc.Config.ServerURL, myHandler.Router())
	if err != nil {
		panic(err)
	}
}
