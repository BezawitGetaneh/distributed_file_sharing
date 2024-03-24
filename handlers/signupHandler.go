// handlers/signup.go
package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var (
	db          *sql.DB
	sessionName = "file_system_session"
	sessionKey  = []byte("your-secret-key")
	store       = sessions.NewCookieStore(sessionKey)
	tmpl        *template.Template
)

// Assume you have the User struct defined in models/user.go
type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error hashing password:", err)
			return
		}

		// Insert user data into MySQL database
		_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", username, string(hashedPassword), "user")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error inserting user into the MySQL database:", err)
			return
		}

		// Insert user data into PostgreSQL database
		_, err = postgresDB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", username, string(hashedPassword), "user")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error inserting user into the PostgreSQL database:", err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err := tmpl.ExecuteTemplate(w, "signup.html", nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
		return
	}
}
