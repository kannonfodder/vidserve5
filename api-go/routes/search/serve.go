package search

import (
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/components"
	"net/http"
	"strconv"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	redgifsapiClient := redgifs.NewClient()
	files, err := redgifsapiClient.Search([]string{query}, 20, page)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}

	redgifs.FormatFileUrls(files)

	w.WriteHeader(http.StatusOK)

	components.Video(files,
		components.More("/search", page)).Render(r.Context(), w)
}
