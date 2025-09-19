package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var apiKey string

func webServerTest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got req")
	io.WriteString(w, "yoyo")
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	fmt.Println(apiKey)
	resp, e := http.Get("http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + title)
	if e == nil {
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
	}
	fmt.Println("Error:", e)
	io.WriteString(w, string(body))
}

func main() {
	godotenv.Load()
	apiKey = os.Getenv("OMDB_KEY")
	http.HandleFunc("/", webServerTest)
	http.HandleFunc("/api/movie", getMovie)
	fmt.Println("starting server at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}
