package creators

import (
	"net/http"

	"kannonfoundry/api-go/auth"
	"kannonfoundry/api-go/feedsvc"

	"github.com/gorilla/mux"
)

func ensureDBReady(w http.ResponseWriter) bool {
	if dbPool == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("database not configured"))
		return false
	}
	return true
}

func writeSubscribed(w http.ResponseWriter, username string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<button id=\"subscribe-btn\" class=\"btn subscribed\" hx-delete=\"/creators/" + username + "/subscribe\" hx-target=\"#subscribe-btn\" hx-swap=\"outerHTML\">Subscribed ✓</button>"))
}

func writeUnsubscribed(w http.ResponseWriter, username string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<button id=\"subscribe-btn\" class=\"btn\" hx-post=\"/creators/" + username + "/subscribe\" hx-target=\"#subscribe-btn\" hx-swap=\"outerHTML\">Subscribe</button>"))
}

// Subscribe creates a creator subscription for the logged-in user.
func Subscribe(w http.ResponseWriter, r *http.Request) {
	if !ensureDBReady(w) {
		return
	}
	user := auth.IsLoggedIn(r)
	if user.IsEmpty() {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Login required"))
		return
	}

	username := mux.Vars(r)["username"]
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing username"))
		return
	}

	// Check if already subscribed; idempotent.
	sub, err := feedsvc.GetSubscriptionByUserAndTerm(dbPool, user.Id, "creator", username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if sub != nil {
		writeSubscribed(w, username)
		return
	}

	_, err = feedsvc.CreateSubscription(dbPool, user.Id, "creator", username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeSubscribed(w, username)
}

// Unsubscribe removes a creator subscription for the logged-in user.
func Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if !ensureDBReady(w) {
		return
	}
	user := auth.IsLoggedIn(r)
	if user.IsEmpty() {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Login required"))
		return
	}

	username := mux.Vars(r)["username"]
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing username"))
		return
	}

	err := feedsvc.DeleteSubscriptionByUserAndTerm(dbPool, user.Id, "creator", username)
	if err != nil {
		// If not found, still return unsubscribed state for idempotency
		w.WriteHeader(http.StatusOK)
		writeUnsubscribed(w, username)
		return
	}

	writeUnsubscribed(w, username)
}

// SubscriptionStatus returns the current button state for HTMX swaps.
func SubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureDBReady(w) {
		return
	}
	user := auth.IsLoggedIn(r)
	username := mux.Vars(r)["username"]
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing username"))
		return
	}

	if user.IsEmpty() {
		// Not logged in: show login prompt button
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<a href=\"/login\" class=\"btn\">Login to subscribe</a>"))
		return
	}

	sub, err := feedsvc.GetSubscriptionByUserAndTerm(dbPool, user.Id, "creator", username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if sub != nil {
		writeSubscribed(w, username)
	} else {
		writeUnsubscribed(w, username)
	}
}
