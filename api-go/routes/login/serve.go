package login

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

// SetDB sets the database pool for the login handler
func SetDB(db *pgxpool.Pool) {
	dbPool = db
}

func render(w http.ResponseWriter, r *http.Request, template templ.Component) {
	layout.Root("Kannonfoundry login", template).Render(r.Context(), w)
}
func Serve(w http.ResponseWriter, r *http.Request) {

	user := auth.IsLoggedIn(r)
	if user.IsEmpty() {
		render(w, r, components.Login(false))
	} else {
		render(w, r, components.Logout(user))
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	// Process login form submission
	username := r.FormValue("username")
	password := r.FormValue("password")

	cookie, err := auth.LoginUserWithDB(dbPool, username, password)
	if err != nil {
		log.Printf("Login failed for user %s: %v", username, err)
		render(w, r, components.Login(true))
		return
	}

	// Set a session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    cookie,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7, // 7 days
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
