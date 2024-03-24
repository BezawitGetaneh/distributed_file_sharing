// handlers/login.go
package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// Include the necessary import for the "text/template" package

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if authenticate(username, password) {
			isAdmin := checkAdmin(username)

			session, _ := store.Get(r, sessionName)
			session.Values["authenticated"] = true

			if isAdmin {
				session.Values["role"] = "admin"
				session.Save(r, w)
				http.Redirect(w, r, "/index", http.StatusSeeOther)
				return
			} else {
				session.Values["role"] = "user"
				session.Save(r, w)
				http.Redirect(w, r, "/view", http.StatusSeeOther)
				return
			}
		}

		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	err := tmpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
		return
	}
}

func authenticate(username, password string) bool {
	// Retrieve the user's hashed password from the database
	row := db.QueryRow("SELECT password FROM users WHERE username = ?", username)
	var hashedPassword string
	err := row.Scan(&hashedPassword)
	if err != nil {
		// If user is not found in MySQL, try retrieving from PostgreSQL
		row = postgresDB.QueryRow("SELECT password FROM users WHERE username = $1", username)
		err = row.Scan(&hashedPassword)
		if err != nil {
			return false
		}
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func checkAdmin(username string) bool {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE username=?", username).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, assuming not an admin
			return false
		}
		log.Println("MySQL Error:", err)
		// MySQL query failed, try PostgreSQL

		// Retrieve role from PostgreSQL
		err = postgresDB.QueryRow("SELECT role FROM users WHERE username=?", username).Scan(&role)
		if err != nil {
			if err == sql.ErrNoRows {
				// User not found, assuming not an admin
				return false
			}
			log.Println("PostgreSQL Error:", err)
			// Error occurred while retrieving role from PostgreSQL, assume not an admin
			return false
		}
	}

	return role == "admin"
}
