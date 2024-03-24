// handlers/downloadHandler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get the file ID from the URL query parameters
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
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set the appropriate headers for downloading
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
	w.Header().Set("Content-Type", http.DetectContentType(file.Data))
	w.Header().Set("Content-Length", strconv.Itoa(len(file.Data)))

	// Write the file content to the response writer
	_, err = w.Write(file.Data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error writing file data to response:", err)
		return
	}
}
