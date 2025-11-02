package login

import (
	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"net/http"

	"github.com/a-h/templ"
)

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
	cookie := auth.LoginUser(username, password)
	if cookie != "" {
		// Set a session cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "session_id",
			Value: cookie,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// If login fails, show the login form again
	render(w, r, components.Login(true))
}
