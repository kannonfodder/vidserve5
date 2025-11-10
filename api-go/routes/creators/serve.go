package creators

import (
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
)

func render(w http.ResponseWriter, r *http.Request, template templ.Component) {
	layout.Root("Kannonfoundry login", template).Render(r.Context(), w)
}
func Serve(w http.ResponseWriter, r *http.Request) {
	//This is where we need to get the username and render the creator page
	vars := mux.Vars(r)
	username := vars["username"]

	redgifsapiClient := redgifs.NewClient()
	files, err := redgifsapiClient.SearchByUser(username, 20, 1)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}

	redgifs.FormatFileUrls(files)

	render(w, r, layout.Creator(username, components.Video(files)))

}
