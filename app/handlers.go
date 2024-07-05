// app/handlers.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
)

// Handler for form submission
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Info("Received a POST request")

	config, err := loadConfig("/app/config/config.json")
	if err != nil {
		http.Error(w, "Could not load config", http.StatusInternalServerError)
		log.Errorf("Error loading config: %v", err)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
		log.Errorf("Error parsing form: %v", err)
		return
	}

	formID := r.FormValue("formid")
	if formID == "" {
		http.Error(w, "Form ID is required", http.StatusBadRequest)
		log.Warn("Form ID is required")
		return
	}

	formConfig, exists := config.Forms[formID]
	if !exists {
		http.Error(w, "Form configuration not found", http.StatusBadRequest)
		log.Warnf("Form configuration not found for ID: %s", formID)
		return
	}

	// Check the referral URL
	referer := r.Referer()
	if referer == "" || !strings.HasPrefix(referer, formConfig.ReferralURL) {
		http.Error(w, "Invalid referral URL", http.StatusForbidden)
		log.Warnf("Invalid referral URL: %s", referer)
		return
	}

	// Check the allowed origins
	origin := r.Header.Get("Origin")
	originAllowed := false
	for _, allowedOrigin := range formConfig.AllowedOrigins {
		if origin == allowedOrigin {
			originAllowed = true
			break
		}
	}
	if !originAllowed {
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		log.Warnf("Origin not allowed: %s", origin)
		return
	}

	// Use bluemonday to create a policy that allows only plain text
	policy := bluemonday.StrictPolicy()

	formData := make(map[string]string)
	for _, field := range formConfig.Fields {
		value := r.FormValue(field.Name)
		if field.Required && strings.TrimSpace(value) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%s is required", field.Name)})
			log.Warnf("%s is required", field.Name)
			return
		}
		if field.MaxLength > 0 && len(value) > field.MaxLength {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%s exceeds maximum length of %d", field.Name, field.MaxLength)})
			log.Warnf("%s exceeds maximum length of %d", field.Name, field.MaxLength)
			return
		}
		formData[field.Name] = policy.Sanitize(value) // Sanitize the input

		if field.Type == "file" {
			if fileHeaders, ok := r.MultipartForm.File[field.Name]; ok {
				for _, fileHeader := range fileHeaders {
					file, err := fileHeader.Open()
					if err != nil {
						http.Error(w, "Could not open uploaded file", http.StatusInternalServerError)
						log.Errorf("Error opening file: %v", err)
						return
					}
					defer file.Close()

					if err := validateFile(fileHeader, field); err != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
						log.Errorf("File validation error: %v", err)
						return
					}

					ext := filepath.Ext(fileHeader.Filename)
					baseName := strings.TrimSuffix(fileHeader.Filename, ext)
					randomString, err := generateRandomString(10)
					if err != nil {
						http.Error(w, "Could not generate random string for file name", http.StatusInternalServerError)
						log.Errorf("Error generating random string: %v", err)
						return
					}
					newFileName := fmt.Sprintf("%s_%s%s", baseName, randomString, ext)

					dst, err := os.Create(fmt.Sprintf("/app/uploads/%s", filepath.Base(newFileName)))
					if err != nil {
						http.Error(w, "Could not create file on server", http.StatusInternalServerError)
						log.Errorf("Error creating file: %v", err)
						return
					}
					defer dst.Close()

					if _, err := io.Copy(dst, file); err != nil {
						http.Error(w, "Could not save file", http.StatusInternalServerError)
						log.Errorf("Error saving file: %v", err)
						return
					}
					formData["file"] = newFileName
					log.Infof("Uploaded file %s saved to /app/uploads/%s", fileHeader.Filename, newFileName)
				}
			}
		}
	}

	db, err := getDB()
	if err != nil {
		log.Errorf("Error opening database: %v", err)
		http.Error(w, "Could not connect to the database", http.StatusInternalServerError)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Errorf("Error beginning transaction: %v", err)
		http.Error(w, "Could not begin database transaction", http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO submissions(form_id, name, email, message, file, read) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Errorf("Error preparing statement: %v", err)
		http.Error(w, "Could not prepare database statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(formID, formData["name"], formData["email"], formData["message"], formData["file"], "N")
	if err != nil {
		tx.Rollback()
		log.Errorf("Error executing statement: %v", err)
		http.Error(w, "Could not execute database statement", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf("Error committing transaction: %v", err)
		http.Error(w, "Could not commit database transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"success": "Form submitted successfully"})
	log.Infof("Form processed successfully, data: %+v", formData)
}

// Handler for user login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		log.Infof("Login attempt with username: %s", username)
		envUsername := os.Getenv("ADMIN_USERNAME")
		envPassword := os.Getenv("ADMIN_PASSWORD")
		if username == envUsername && password == envPassword {
			session.Values["authenticated"] = true
			session.Options = &sessions.Options{
				Path:     "/",
				HttpOnly: true,
				MaxAge:   3600, // 1 hour
			}
			err := session.Save(r, w)
			if err != nil {
				log.Errorf("Error saving session: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			log.Infof("User %s authenticated successfully", username)
			http.Redirect(w, r, "/submissions", http.StatusFound)
			return
		}
		log.Warnf("Invalid login attempt for username: %s", username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	http.ServeFile(w, r, "/app/backend/login.html")
}

// Handler for user logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		log.Errorf("Error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Info("User logged out")
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Handler to view submissions (admin)
func viewSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/app/backend/index.html")
}

// API handler to fetch submissions
func apiSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := getDB()
	if err != nil {
		log.Errorf("Error opening database: %v", err)
		http.Error(w, "Could not connect to the database", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, form_id, name, email, message, file, read, created_at FROM submissions")
	if err != nil {
		log.Errorf("Error querying database: %v", err)
		http.Error(w, "Could not query the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var submissions []map[string]interface{}
	for rows.Next() {
		var id int
		var formID, name, email, message, file, read, createdAt string
		err := rows.Scan(&id, &formID, &name, &email, &message, &file, &read, &createdAt)
		if err != nil {
			log.Errorf("Error scanning row: %v", err)
			http.Error(w, "Could not read data from the database", http.StatusInternalServerError)
			return
		}
		submission := map[string]interface{}{
			"id":         id,
			"form_id":    formID,
			"name":       name,
			"email":      email,
			"message":    message,
			"file":       file,
			"read":       read,
			"created_at": createdAt,
		}
		submissions = append(submissions, submission)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// Handler to delete a submission by ID (admin)
func deleteSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	db, err := getDB()
	if err != nil {
		log.Errorf("Error opening database: %v", err)
		http.Error(w, "Could not connect to the database", http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("DELETE FROM submissions WHERE id = ?")
	if err != nil {
		log.Errorf("Error preparing delete statement: %v", err)
		http.Error(w, "Could not prepare delete statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		log.Errorf("Error executing delete statement: %v", err)
		http.Error(w, "Could not delete submission", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Infof("Submission with ID %s deleted", id)
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// API handler to fetch rate limits
func apiRateLimitsHandler(w http.ResponseWriter, r *http.Request) {
	rateLimits := getRateLimits()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateLimits)
}

func getRateLimits() map[string][]time.Time {
	rateLimits := make(map[string][]time.Time)
	rateLimiter.mu.Lock()
	defer rateLimiter.mu.Unlock()

	for ip, visitor := range rateLimiter.visitors {
		visitor.mu.Lock()
		rateLimits[ip] = append([]time.Time(nil), visitor.timestamps...)
		visitor.mu.Unlock()
	}
	return rateLimits
}

// API handler to clear rate limits for a specific IP
func clearRateLimitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ip := vars["ip"]

	rateLimiter.mu.Lock()
	delete(rateLimiter.visitors, ip)
	rateLimiter.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
	log.Infof("Rate limits cleared for IP %s", ip)
}
