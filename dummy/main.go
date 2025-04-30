// Package main contains the logic for dummy backend.
package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("the port is not set")
	}

	mux := chi.NewMux()

	mux.Use(
		middleware.RequestID,
		middleware.Heartbeat("/health"),
		middleware.Logger,
	)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handled request successfully"))
	})

	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("starting dummy backend on %s", srv.Addr)

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("failed to start the server: " + err.Error())
	}
}
