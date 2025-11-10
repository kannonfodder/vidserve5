package search

import (
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/components"
	"net/http"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")
	redgifsapiClient := redgifs.NewClient()
	files, err := redgifsapiClient.Search([]string{query}, 20, 1)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}

	redgifs.FormatFileUrls(files)

	w.WriteHeader(http.StatusOK)
	components.Video(files).Render(r.Context(), w)
}
