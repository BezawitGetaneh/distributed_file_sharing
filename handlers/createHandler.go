// handlers/createHandler.go
package handlers

import (
	"fmt"
	"io"
	"net/http"
)

var (
	createFile = make(chan int)
)

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20) // Limit the file size to 10MB

		fileName := r.FormValue("fileName")
		file, _, err := r.FormFile("fileContent")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error retrieving file from form:", err)
			return
		}
		defer file.Close()

		fileContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error reading file content:", err)
			return
		}

		// Create a new file entry in the database
		result, err := db.Exec("INSERT INTO files (name, content, data) VALUES (?, ?, ?)", fileName, string(fileContent), fileContent)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error inserting file into the database:", err)
			return
		}
		// Insert into the second database (PostgreSQL)
		_, err = postgresDB.Exec("INSERT INTO files (name, content, data) VALUES ($1, $2, $3)", fileName, string(fileContent), fileContent)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error inserting file into the second database (PostgreSQL):", err)
			return
		}
		// Retrieve the generated ID of the new file
		fileID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error retrieving last insert ID:", err)
			return
		}

		// Send the new file ID to the createFile channel
		createFile <- int(fileID)

		// Redirect to the edit page for the newly created file
		http.Redirect(w, r, fmt.Sprintf("/index"), http.StatusSeeOther)
		return
	}

	err := tmpl.ExecuteTemplate(w, "upload.html", nil)
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
