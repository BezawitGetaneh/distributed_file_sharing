// handlers/indexhandler.go
package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)

var (
	postgresDB *sql.DB
	files      = make(map[int]*File)
	mutex      sync.Mutex

	createTableSQL = `
    CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTO_INCREMENT,
        name TEXT,
        content TEXT
    );
    `
)

type Edit struct {
	ID        int
	Document  string
	User      string
	Content   string
	Timestamp time.Time
	// Add other edit-related fields as needed
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve all files from the database
	rows, err := db.Query("SELECT id, name, content FROM files")
	if err != nil {
		// Try retrieving files from the PostgreSQL database
		postgresRows, err := postgresDB.Query("SELECT id, name, content FROM files")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error retrieving files from PostgreSQL:", err)
			return
		}
		defer postgresRows.Close()

		// Create a slice to store the files
		files := make([]*File, 0)

		// Iterate over the PostgreSQL rows and populate the files slice
		for postgresRows.Next() {
			file := &File{}
			err := postgresRows.Scan(&file.ID, &file.Name, &file.Content)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Error scanning file row from PostgreSQL:", err)
				return
			}
			files = append(files, file)
		}
		err = postgresRows.Err()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error iterating over file rows from PostgreSQL:", err)
			return
		}

		// Pass the files data to the template for rendering
		data := struct {
			Files []*File
		}{
			Files: files,
		}

		err = tmpl.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error executing template:", err)
			return
		}
	} else {
		defer rows.Close()

		// Create a slice to store the files
		files := make([]*File, 0)

		// Iterate over the rows and populate the files slice
		for rows.Next() {
			file := &File{}
			err := rows.Scan(&file.ID, &file.Name, &file.Content)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Error scanning file row:", err)
				return
			}
			files = append(files, file)
		}
		err = rows.Err()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error iterating over file rows:", err)
			return
		}

		// Pass the files data to the template for rendering
		data := struct {
			Files []*File
		}{
			Files: files,
		}

		err = tmpl.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error executing template:", err)
			return
		}
		session, _ := store.Get(r, sessionName)
		authenticated := session.Values["authenticated"]
		if authenticated == nil || authenticated.(bool) != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}
}
