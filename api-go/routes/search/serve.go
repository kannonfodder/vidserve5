package search

import (
	"kannonfoundry/api-go/api"
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/components"
	"net/http"
	"strings"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("search")
	redgifsapiClient := redgifs.NewClient()
	files, err := redgifsapiClient.Search([]string{query})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Error during search: " + err.Error()))
		return
	}
	formatFileUrls(files)

	w.WriteHeader(http.StatusOK)
	components.Video(files).Render(r.Context(), w)
}
func formatFileUrls(files []api.FileToSend) {
	for i, file := range files {
		files[i].URL = strings.Replace(file.URL, "https://media.redgifs.com/", "/rgp/", 1)
	}
}
