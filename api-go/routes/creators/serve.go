package creators

import (
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
)

func render(w http.ResponseWriter, r *http.Request, template templ.Component, username string) {
	layout.Root(username, template).Render(r.Context(), w)
}
func Serve(w http.ResponseWriter, r *http.Request) {
	//This is where we need to get the username and render the creator page
	username := mux.Vars(r)["username"]
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	redgifsapiClient := redgifs.NewClient()
	files, err := redgifsapiClient.SearchByUser(username, 20, page)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}

	redgifs.FormatFileUrls(files)
	if page == 1 {
		render(w, r,
			layout.Creator(username, components.Video(files,
				components.More("/creators/"+username, page, ""))), username)
	} else {
		components.Video(files,
			components.More("/creators/"+username, page, "")).Render(r.Context(), w)
	}

}
