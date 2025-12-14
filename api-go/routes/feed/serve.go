package feed

import (
	"net/http"
	"strconv"

	"kannonfoundry/api-go/api"
	"kannonfoundry/api-go/api/redgifs"
	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/components"
	"kannonfoundry/api-go/components/layout"
	"kannonfoundry/api-go/feedsvc"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func SetDB(pool *pgxpool.Pool) {
	dbPool = pool
}

// Serve renders the authenticated user's feed with pagination.
func Serve(w http.ResponseWriter, r *http.Request) {
	if dbPool == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("database not configured"))
		return
	}

	user := auth.IsLoggedIn(r)
	if user.IsEmpty() {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	const limit = 20
	offset := (page - 1) * limit

	items, err := feedsvc.GetUserFeed(dbPool, user.Id, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error fetching feed: " + err.Error()))
		return
	}

	// Map to api.FileToSend and proxy URLs
	var files []api.FileToSend
	for _, it := range items {
		files = append(files, api.FileToSend{
			Name:     it.VideoId,
			URL:      redgifs.ProxyURL(it.Url),
			Username: it.Username,
		})
	}

	content := components.Video(files, components.More("/feed", page, ""))

	if page == 1 {
		layout.Root("Feed", layout.Feed(user, content)).Render(r.Context(), w)
	} else {
		content.Render(r.Context(), w)
	}
}
