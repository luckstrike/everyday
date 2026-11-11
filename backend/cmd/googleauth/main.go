// Command googleauth signs in to Google once and saves the result for the server.
//
//	make google-auth ARGS="-client-file ~/Downloads/client_secret.json"
//
// The OAuth client ID and secret come from -client-file, or from the
// EVERYDAY_GOOGLE_CLIENT_ID and EVERYDAY_GOOGLE_CLIENT_SECRET variables, or from an
// earlier run's saved file, so the secret is never typed on the command line.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"everyday/backend/internal/gcal"
)

func main() {
	clientFile := flag.String("client-file", "", "JSON file with the OAuth client_id and client_secret")
	credentialsPath := flag.String("credentials", os.Getenv("EVERYDAY_GOOGLE_CREDENTIALS"), "where to save the credentials (default: your user config directory)")
	flag.Parse()

	if *credentialsPath == "" {
		path, err := gcal.DefaultCredentialsPath()
		if err != nil {
			log.Fatal(err)
		}
		*credentialsPath = path
	}

	// Start from any saved client so signing in again needs no flags.
	creds, err := gcal.LoadCredentials(*credentialsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal(err)
	}
	switch {
	case *clientFile != "":
		if creds.ClientID, creds.ClientSecret, err = gcal.ReadClientFile(*clientFile); err != nil {
			log.Fatal(err)
		}
	case os.Getenv("EVERYDAY_GOOGLE_CLIENT_ID") != "" && os.Getenv("EVERYDAY_GOOGLE_CLIENT_SECRET") != "":
		creds.ClientID = os.Getenv("EVERYDAY_GOOGLE_CLIENT_ID")
		creds.ClientSecret = os.Getenv("EVERYDAY_GOOGLE_CLIENT_SECRET")
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		log.Fatal("no OAuth client found: pass -client-file, or set EVERYDAY_GOOGLE_CLIENT_ID and EVERYDAY_GOOGLE_CLIENT_SECRET")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	oauth := &gcal.OAuth{ClientID: creds.ClientID, ClientSecret: creds.ClientSecret}
	creds.RefreshToken, err = oauth.Authorize(ctx, func(authURL string) {
		fmt.Println("Open this URL in your browser to sign in with Google:")
		fmt.Println()
		fmt.Println(authURL)
		fmt.Println()
		if err := openBrowser(authURL); err != nil {
			fmt.Println("(Could not open a browser automatically; copy the URL above.)")
		}
		fmt.Println("Waiting for you to approve access…")
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := gcal.SaveCredentials(*credentialsPath, creds); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Signed in. Saved credentials to %s\nRestart the dashboard server to pick them up.\n", *credentialsPath)
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	}
	return errors.New("unsupported platform")
}
