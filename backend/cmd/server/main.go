package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"everyday/backend/internal/api"
	"everyday/backend/internal/config"
	"everyday/backend/internal/gcal"
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
		if calendar := connectGoogleCalendar(ctx, cfg); calendar != nil {
			opts.Calendar = calendar
		}
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

// connectGoogleCalendar starts reading Google Calendar once the user has signed in
// with `make google-auth`. Until then it returns nil and the dashboard's calendar
// widget says it is not connected. A damaged credentials file stops the server,
// like any other bad setting.
func connectGoogleCalendar(ctx context.Context, cfg config.Config) *gcal.Service {
	path := cfg.GoogleCredentials
	if path == "" {
		var err error
		if path, err = gcal.DefaultCredentialsPath(); err != nil {
			log.Fatal(err)
		}
	}
	creds, err := gcal.LoadCredentials(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		log.Printf("Google Calendar is not connected (no %s); run make google-auth", path)
		return nil
	case err != nil:
		log.Fatal(err)
	case creds.RefreshToken == "":
		log.Printf("Google Calendar is not connected (%s has no sign-in yet); run make google-auth", path)
		return nil
	}
	oauth := &gcal.OAuth{ClientID: creds.ClientID, ClientSecret: creds.ClientSecret}
	service := gcal.NewService(&gcal.Client{
		Tokens:      gcal.NewRefreshTokens(oauth, creds.RefreshToken),
		CalendarIDs: cfg.GoogleCalendars,
	})
	go service.Run(ctx)
	log.Printf("Serving Google Calendar events from %d calendar(s)", len(cfg.GoogleCalendars))
	return service
}
