package register

import (
	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

// SetDB sets the database pool for the registration handler
func SetDB(db *pgxpool.Pool) {
	dbPool = db
}

func render(w http.ResponseWriter, r *http.Request, template templ.Component) {
	layout.Root("Register - Kannonfoundry", template).Render(r.Context(), w)
}

func Serve(w http.ResponseWriter, r *http.Request) {
	// Check if already logged in
	user := auth.IsLoggedIn(r)
	if !user.IsEmpty() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// Show registration form
	render(w, r, components.Register(false, ""))
}

func RegisterPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate input
	if username == "" || password == "" {
		render(w, r, components.Register(true, "Username and password are required"))
		return
	}

	if password != confirmPassword {
		render(w, r, components.Register(true, "Passwords do not match"))
		return
	}

	if len(password) < 8 {
		render(w, r, components.Register(true, "Password must be at least 8 characters"))
		return
	}

	// Create user
	user, err := auth.CreateUser(dbPool, username, password)
	if err != nil {
		log.Printf("Registration failed for user %s: %v", username, err)
		render(w, r, components.Register(true, "Username already exists or registration failed"))
		return
	}

	log.Printf("User %s registered successfully", user.Username)

	// Auto-login after registration
	cookie, err := auth.CreateUserCookie(*user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    cookie,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
