package creators

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	//This is where we need to get the username and render the creator page
	vars := mux.Vars(r)
	username := vars["username"]
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Creator page for user: " + username))
}
