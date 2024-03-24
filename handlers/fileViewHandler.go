// handlers/fileViewHandler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func FileViewHandler(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the database
	row := db.QueryRow("SELECT id, name, content FROM files WHERE id = ?", fileID)
	file := &File{}
	err = row.Scan(&file.ID, &file.Name, &file.Content)
	if err != nil {
		// Try retrieving the file from the PostgreSQL database if not found in MySQL
		postgresRow := postgresDB.QueryRow("SELECT id, name, content FROM files WHERE id = $1", fileID)
		err = postgresRow.Scan(&file.ID, &file.Name, &file.Content)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	// Render the file content in a read-only format
	data := struct {
		File *File
	}{
		File: file,
	}

	err = tmpl.ExecuteTemplate(w, "file_view.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
		return
	}
}
