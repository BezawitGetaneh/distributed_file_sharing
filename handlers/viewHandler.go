// handlers/viewHandler.go
package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
)

func init() {
	// Open the MySQL database
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/file_system")
	if err != nil {
		fmt.Println("Error opening the database:", err)
		return
	}
	// Open the PostgreSQL database
	postgresDB, err = sql.Open("postgres", "host=127.0.0.1 port=5432 user=postgres password=root dbname=file_system sslmode=disable")
	if err != nil {
		fmt.Println("Error opening the PostgreSQL database:", err)
		return
	}
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // Session expires in 24 hours
		HttpOnly: true,
	}
	// Create files table if not exists
	_, err = db.Exec(createTableSQL)
	if err != nil {
		fmt.Println("Error creating the table:", err)
		return
	}

	// Parse HTML templates
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve all files from the database
	rows, err := db.Query("SELECT id, name, content FROM files")
	if err != nil {
		// MySQL server is unreachable, fallback to PostgreSQL
		fmt.Println("Error retrieving files from MySQL:", err)

		// Retrieve all files from PostgreSQL
		pgRows, pgErr := postgresDB.Query("SELECT id, name, content FROM files")
		if pgErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error retrieving files from PostgreSQL:", pgErr)
			return
		}
		defer pgRows.Close()

		// Create a slice to store the files
		files := make([]*File, 0)

		// Iterate over the PostgreSQL rows and populate the files slice
		for pgRows.Next() {
			file := &File{}
			err := pgRows.Scan(&file.ID, &file.Name, &file.Content)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Error scanning PostgreSQL file row:", err)
				return
			}
			files = append(files, file)
		}
		pgErr = pgRows.Err()
		if pgErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error iterating over PostgreSQL file rows:", pgErr)
			return
		}

		// Pass the files data to the template for rendering
		data := struct {
			Files []*File
		}{
			Files: files,
		}

		err := tmpl.ExecuteTemplate(w, "view.html", data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error executing template:", err)
			return
		}
		return
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

		err = tmpl.ExecuteTemplate(w, "view.html", data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error executing template:", err)
			return
		}
	}
}
