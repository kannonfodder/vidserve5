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
	newUrl := "https://media.redgifs.com/" + r.URL.Path
	sfw := os.Getenv("SFW")
	if sfw == "true" {
		newUrl = "https://c959e687-9816-4324-9b78-6e34277b9c81.mdnplay.dev/shared-assets/videos/flower.mp4"
	}
	req, err := http.NewRequest("GET", newUrl, nil)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
