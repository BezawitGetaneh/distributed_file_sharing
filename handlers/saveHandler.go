// handlers/saveHandler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

var updateFile = make(chan File)

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()

		fileIDStr := r.FormValue("fileID")
		fileID, err := strconv.Atoi(fileIDStr)
		if err != nil {
			http.Error(w, "Invalid file ID", http.StatusBadRequest)
			return
		}

		fileContent := r.FormValue("fileContent")
		// Start a transaction for MySQL
		mysqlTx, err := db.Begin()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error starting MySQL transaction:", err)
			return
		}

		// Start a transaction for PostgreSQL
		pgTx, err := postgresDB.Begin()
		if err != nil {
			mysqlTx.Rollback()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error starting PostgreSQL transaction:", err)
			return
		}

		// Update the file content in MySQL
		_, err = mysqlTx.Exec("UPDATE files SET content = ? WHERE id = ?", fileContent, fileID)
		if err != nil {
			mysqlTx.Rollback()
			pgTx.Rollback()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error updating file content in MySQL:", err)
			return
		}

		// Update the file content in PostgreSQL
		_, err = pgTx.Exec("UPDATE files SET content = $1 WHERE id = $2", fileContent, fileID)
		if err != nil {
			mysqlTx.Rollback()
			pgTx.Rollback()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error updating file content in PostgreSQL:", err)
			return
		}

		// Commit the transactions
		err = mysqlTx.Commit()
		if err != nil {
			pgTx.Rollback()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error committing MySQL transaction:", err)
			return
		}

		err = pgTx.Commit()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error committing PostgreSQL transaction:", err)
			return
		}

		// Retrieve the updated file from MySQL
		mysqlRow := db.QueryRow("SELECT id, name, content, data FROM files WHERE id = ?", fileID)
		updatedFile := &File{}
		err = mysqlRow.Scan(&updatedFile.ID, &updatedFile.Name, &updatedFile.Content, &updatedFile.Data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error retrieving updated file from MySQL:", err)
			return
		}

		updateFile <- *updatedFile

		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
