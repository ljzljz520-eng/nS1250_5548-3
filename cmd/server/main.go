package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"time"

	"warehouse-workers/internal/account"
	"warehouse-workers/internal/httpapi"
)

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	repository := account.NewMemoryRepository([]account.Worker{
		{
			Name:       "Lin Wei",
			EmployeeID: "WH-1001",
			Phone:      "+8613812345678",
			Email:      "lin.wei@example.com",
			Team:       "Inbound A",
			Status:     account.StatusActive,
		},
	})
	service := account.NewService(repository)
	server := &http.Server{
		Addr:              *address,
		Handler:           httpapi.New(service).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("warehouse worker API listening on %s", *address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
