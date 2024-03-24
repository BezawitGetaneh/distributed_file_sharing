// handlers/logoutHandler.go
package handlers

import (
	"net/http"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	session.Values["authenticated"] = false
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
