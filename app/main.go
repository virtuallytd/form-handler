// app/main.go
package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var rateLimiter *RateLimiter

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("/app/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Create uploads directory if it doesn't exist
	if _, err := os.Stat("/app/uploads"); os.IsNotExist(err) {
		if err := os.Mkdir("/app/uploads", os.ModePerm); err != nil {
			log.Fatalf("Could not create uploads directory: %v", err)
		}
	}

	// Load the application configuration
	configPath := "/app/config/config.json"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize the database
	initDatabase()

	// Initialize the rate limiter
	rateLimiter = newRateLimiter()

	// Create a new router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("GET")
	r.Handle("/submissions", authMiddleware(http.HandlerFunc(viewSubmissionsHandler)))
	r.Handle("/api/submissions", authMiddleware(http.HandlerFunc(apiSubmissionsHandler))).Methods("GET")
	r.Handle("/api/submissions/{id}", authMiddleware(http.HandlerFunc(deleteSubmissionHandler))).Methods("DELETE")
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.Handle("/api/rate-limits", authMiddleware(http.HandlerFunc(apiRateLimitsHandler))).Methods("GET")
	r.Handle("/rate-limits", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/app/backend/rate_limits.html")
	}))).Methods("GET")
	r.Handle("/api/rate-limits/{ip}", authMiddleware(http.HandlerFunc(clearRateLimitHandler))).Methods("DELETE")
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("/app/uploads/"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/app/backend/static/"))))

	// Apply rate limit and CORS middleware to form submission route
	submitHandler := rateLimitMiddleware(http.HandlerFunc(formHandler), rateLimiter, config)
	submitHandler = dynamicCORSMiddleware(submitHandler, config)
	r.Handle("/api/forms", submitHandler).Methods("POST")

	// Start the server
	log.Info("Server started at :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
