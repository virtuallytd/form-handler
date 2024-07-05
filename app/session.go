package main

import (
	"os"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

// Initialize the session store
func init() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET environment variable is not set")
	}
	store = sessions.NewCookieStore([]byte(secret))
}
