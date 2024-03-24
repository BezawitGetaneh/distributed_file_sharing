package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

type File struct {
	ID      int
	Name    string
	Content string
	Data    []byte
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the database
	row := db.QueryRow("SELECT id, name, content, data FROM files WHERE id = ?", fileID)
	file := &File{}
	err = row.Scan(&file.ID, &file.Name, &file.Content, &file.Data)
	if err != nil {
		// File not found in MySQL, try retrieving from PostgreSQL
		pgRow := postgresDB.QueryRow("SELECT id, name, content, data FROM files WHERE id = $1", fileID)
		file := &File{}
		err = pgRow.Scan(&file.ID, &file.Name, &file.Content, &file.Data)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	err = tmpl.ExecuteTemplate(w, "edit.html", file)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
		return
	}
}
