package rgp

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	w.WriteHeader(http.StatusOK)

	// Extract host from query parameter (defaults to "media")
	host := r.URL.Query().Get("h")
	if host == "" {
		host = "media"
	}

	// Map short host names to full domain
	hostMap := map[string]string{
		"media":   "media.redgifs.com",
		"thumbs":  "thumbs.redgifs.com",
		"userpic": "userpic.redgifs.com",
		"i":       "i.redgifs.com",
	}

	fullHost, ok := hostMap[host]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid host: " + host))
		return
	}

	// Construct full URL from path and host
	path := r.URL.Path
	if path[0] == '/' {
		path = path[1:]
	}

	newUrl := "https://" + fullHost + "/" + path

	sfw := os.Getenv("SFW")
	if sfw == "true" {
		newUrl = "https://c959e687-9816-4324-9b78-6e34277b9c81.mdnplay.dev/shared-assets/videos/flower.mp4"
	}

	fmt.Println("Proxying to: " + newUrl)
	req, err := http.NewRequest("GET", newUrl, nil)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
