// Package main contains the logic for dummy backend.
package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("the port is not set")
	}

	mode := os.Getenv("MODE")

	mux := chi.NewMux()

	mux.Use(
		middleware.RequestID,
		middleware.Heartbeat("/health"),
		middleware.Logger,
	)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if mode == "slow" {
			time.Sleep(time.Millisecond * 500)
		}

		type response struct {
			Result string `json:"result"`
		}

		resp := response{
			Result: "handled request successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
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
