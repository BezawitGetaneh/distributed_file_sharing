// handlers/deleteHandler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var broadcast = make(chan Message)

type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get the file ID from the form
		fileIDStr := r.FormValue("fileID")
		fileID, err := strconv.Atoi(fileIDStr)
		if err != nil || fileID <= 0 {
			http.Error(w, "Invalid file ID", http.StatusBadRequest)
			return
		}

		// Use a wait group to ensure both deletions are completed before proceeding
		var wg sync.WaitGroup
		wg.Add(2)

		// Delete the file from MySQL
		go func() {
			_, err := db.Exec("DELETE FROM files WHERE id = ?", fileID)
			if err != nil {
				fmt.Println("Error deleting file from MySQL:", err)
			}
			wg.Done()
		}()

		// Delete the file from PostgreSQL
		go func() {
			// Log the SQL query for debugging
			sqlQuery := "DELETE FROM files WHERE id = $1"
			_, err := postgresDB.Exec(sqlQuery, fileID)
			fmt.Println("Deleting fileID:", fileID)
			fmt.Println("SQL Query:", sqlQuery)
			if err != nil {
				fmt.Println("Error deleting file from PostgreSQL:", err)
				fmt.Println("SQL Query:", sqlQuery)
			}
			wg.Done()
		}()

		wg.Wait() // Wait for both deletions to complete

		// Broadcast a deletion message to all connected clients
		broadcast <- Message{
			Action:  "file_delete",
			Content: strconv.Itoa(fileID),
		}

		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
