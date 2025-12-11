package main

import (
	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/components/layout"
	"kannonfoundry/api-go/db"
	"kannonfoundry/api-go/feedsvc"
	"kannonfoundry/api-go/routes/creators"
	"kannonfoundry/api-go/routes/login"
	"kannonfoundry/api-go/routes/logout"
	"kannonfoundry/api-go/routes/register"
	"kannonfoundry/api-go/routes/rgp"
	"kannonfoundry/api-go/routes/search"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env", ".key.env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Initialize database connection
	dbPool, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// Set database for routes that need it
	login.SetDB(dbPool)
	register.SetDB(dbPool)

	// Start background worker for feed updates
	go feedsvc.StartWorker(dbPool)

	r := mux.NewRouter()
	log.Println("Server started on :8080")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		guid := uuid.New().String()
		os.Create("files/" + guid)

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Location", "/files/"+guid)
		w.Write([]byte("Done"))
	})
	r.PathPrefix("/rgp/").Handler(http.StripPrefix("/rgp/", http.HandlerFunc(rgp.Serve)))
	r.HandleFunc("/search", search.Serve)
	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("files"))))
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	r.HandleFunc("/creators/{username}", creators.Serve)
	r.HandleFunc("/login", login.LoginPost).Methods("POST")
	r.HandleFunc("/login", login.Serve).Methods("GET")
	r.HandleFunc("/register", register.RegisterPost).Methods("POST")
	r.HandleFunc("/register", register.Serve).Methods("GET")
	r.HandleFunc("/logout", logout.Serve)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user := auth.IsLoggedIn(r)
		layout.Root("Kannonfoundry", layout.Search(user)).Render(r.Context(), w)
	})
	log.Fatal(srv.ListenAndServe())
}
