package main

import (
	"DS_FileSystem/handlers"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var (
	sessionName = "file_system_session"
	sessionKey  = []byte("your-secret-key")
	store       = sessions.NewCookieStore(sessionKey)
	updateFile  = make(chan File)
	createFile  = make(chan int)
)
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var (
	db             *sql.DB
	postgresDB     *sql.DB
	files          = make(map[int]*File)
	mutex          sync.Mutex
	tmpl           *template.Template
	createTableSQL = `
    CREATE TABLE IF NOT EXISTS files (
        id INTEGER PRIMARY KEY AUTO_INCREMENT,
        name TEXT,
        content TEXT
    );
    `
)

type User struct {
	ID       int
	Username string
	password string
}
type Message struct {
	Action  string `json:"action"`
	Content string `json:"content"`
}
type File struct {
	ID      int
	Name    string
	Content string
	Data    []byte
}
type Edit struct {
	ID        int
	Document  string
	User      string
	Content   string
	Timestamp time.Time
	// Add other edit-related fields as needed
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// func init() {
// 	// Open the MySQL database
// 	var err error
// 	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/file_system")
// 	if err != nil {
// 		fmt.Println("Error opening the database:", err)
// 		return
// 	}
// 	// Open the PostgreSQL database
// 	postgresDB, err = sql.Open("postgres", "host=127.0.0.1 port=5432 user=postgres password=root dbname=file_system sslmode=disable")
// 	if err != nil {
// 		fmt.Println("Error opening the PostgreSQL database:", err)
// 		return
// 	}
// 	store.Options = &sessions.Options{
// 		Path:     "/",
// 		MaxAge:   86400, // Session expires in 24 hours
// 		HttpOnly: true,
// 	}
// 	// Create files table if not exists
// 	_, err = db.Exec(createTableSQL)
// 	if err != nil {
// 		fmt.Println("Error creating the table:", err)
// 		return
// 	}

// 	// Parse HTML templates
// 	tmpl = template.Must(template.ParseGlob("templates/*.html"))
// }

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	log.Println("WebSocket connection closed.")

	// Register new client
	clients[conn] = true
	log.Println("New WebSocket connection registered.")

	// Listen for incoming messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading JSON:", err)
			delete(clients, conn)
			// conn.Close()
			break
		}
		log.Printf("Received WebSocket message: %+v\n", msg)
		broadcast <- msg
		log.Println("Message broadcasted.")
	}
}

func main() {
	go func() {
		for {
			select {
			case msg := <-broadcast:
				// Broadcast regular WebSocket messages
				for client := range clients {
					err := client.WriteJSON(msg)
					if err != nil {
						log.Println(err)
						client.Close()
						delete(clients, client)
					}
				}
			case updatedFile := <-updateFile:
				// Broadcast file update messages
				for client := range clients {
					err := client.WriteJSON(Message{
						Action:  "file_update",
						Content: updatedFile.Content,
					})
					if err != nil {
						log.Println(err)
						client.Close()
						delete(clients, client)
					}
				}
			case newFileID := <-createFile:
				// Broadcast file creation messages
				for client := range clients {
					err := client.WriteJSON(Message{
						Action:  "file_create",
						Content: strconv.Itoa(newFileID),
					})
					if err != nil {
						log.Println(err)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}
	}()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	http.HandleFunc("/create", handlers.CreateHandler)
	http.HandleFunc("/edit", handlers.EditHandler)
	http.HandleFunc("/save", handlers.SaveHandler)
	http.HandleFunc("/delete", handlers.DeleteHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/signup", handlers.SignupHandler)
	http.HandleFunc("/view", handlers.ViewHandler)
	http.HandleFunc("/index", handlers.IndexHandler)
	http.HandleFunc("/file/view", handlers.FileViewHandler)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/download", handlers.DownloadHandler)

	fmt.Println("Server is running on http://localhost:8081")
	http.ListenAndServe("localhost:8081", nil)
}
