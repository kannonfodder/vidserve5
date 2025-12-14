package creators

import (
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"kannonfoundry/api-go/feedsvc"
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
	creator, err := redgifsapiClient.GetCreator(username)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error fetching creator: " + err.Error()))
		return
	}
	// Proxy the profile image URL to avoid external 403s
	creator.ProfileImageUrl = redgifs.ProxyURL(creator.ProfileImageUrl)
	files, err := redgifsapiClient.SearchByUser(username, 20, page)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}

	// Auth + subscription state
	user := auth.IsLoggedIn(r)
	isLoggedIn := !user.IsEmpty()
	isSubscribed := false
	if isLoggedIn && dbPool != nil {
		if sub, err := feedsvc.GetSubscriptionByUserAndTerm(dbPool, user.Id, "creator", username); err == nil && sub != nil {
			isSubscribed = true
		}
	}

	redgifs.FormatFileUrls(files)
	if page == 1 {
		render(w, r,
			layout.Creator(*creator, components.Video(files,
				components.More("/creators/"+username, page, "")), isLoggedIn, isSubscribed), username)
	} else {
		components.Video(files,
			components.More("/creators/"+username, page, "")).Render(r.Context(), w)
	}

}
