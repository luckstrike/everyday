package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"everyday/backend/internal/api"
	"everyday/backend/internal/config"
	"everyday/backend/internal/weather"
)

func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	opts := api.Options{Mock: cfg.Mock, Location: cfg.Location.Name}
	if cfg.Mock {
		log.Printf("Serving mock data (EVERYDAY_MOCK is set)")
	} else {
		service := weather.NewService(&weather.Client{Location: cfg.Location})
		go service.Run(ctx)
		opts.Weather = service
		log.Printf("Serving live weather for %s", cfg.Location.Name)
	}

	server := &http.Server{Addr: "127.0.0.1:8080", Handler: api.New(opts), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	log.Printf("Dashboard API at http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
