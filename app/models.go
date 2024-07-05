// app/models.go
package main

import (
	"net/http"

	_ "github.com/mattn/go-sqlite3" // Ensure the SQLite3 driver is imported
)

// Initialize the database and create the submissions table if it doesn't exist
func initDatabase() {
	db, err := getDB()
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS submissions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        form_id TEXT,
        name TEXT,
        email TEXT,
        message TEXT,
        file TEXT,
        read TEXT DEFAULT 'N',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

// Handle the deletion of a submission by ID
func handleDeleteSubmission(w http.ResponseWriter, r *http.Request, id string) {
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
