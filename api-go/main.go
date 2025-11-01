package main

import (
	"fmt"
	"html/template"
	"io"
	"kannonfoundry/api-go/api"
	"kannonfoundry/api-go/api/redgifs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", serveTemplate)

	log.Println("Server started on :8080")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		guid := uuid.New().String()
		os.Create("files/" + guid)

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Location", "/files/"+guid)
		w.Write([]byte("Done"))
	})
	r.PathPrefix("/rgp/").Handler(http.StripPrefix("/rgp/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		w.WriteHeader(http.StatusOK)
		newUrl := "https://media.redgifs.com/" + r.URL.Path
		fmt.Println("Fetching from redgifs: " + newUrl)
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
		fmt.Println("Response status: " + resp.Status)
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	})))
	r.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.FormValue("search")
		redgifsapiClient := redgifs.NewClient()
		files, err := redgifsapiClient.Search([]string{query})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Error during search: " + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Search results: " + formatFiles(files)))
	})

	r.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir("files"))))
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func formatFiles(files []api.FileToSend) string {
	result := "<div class=\"row\">"
	for _, file := range files {
		result += fmt.Sprintf("<div class=\"col-12 col-md-4 mb-2\"><video class=\"w-100\" src=\"%s\" controls></video></div>", strings.Replace(file.URL, "https://media.redgifs.com/", "/rgp/", 1))
	}
	result += "</div>"
	return result
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	layout := filepath.Join("assets", "layout.html")
	file := filepath.Join("assets", filepath.Clean(r.URL.Path))
	fmt.Println(file)
	if file == "assets" {
		file = filepath.Join("assets", "index.html")
	}

	tmpl, err := template.ParseFiles(layout, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout", nil)
}
