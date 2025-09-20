package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var apiKey string

func webServerTest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got req")
	io.WriteString(w, "yoyo")
}

type MovieResponse struct {
	Title    string        `json:"Title"`
	Year     string        `json:"Year"`
	Plot     string        `json:"Plot"`
	Country  string        `json:"Country"`
	Awards   string        `json:"Awards"`
	Director string        `json:"Director"`
	Ratings  []interface{} `json:"Ratings"`
}

type EpisodeResponse struct {
	Title      string `json:"Title"`
	Released   string `json:"Released"`
	Season     string `json:"Season"`
	Episode    string `json:"Episode"`
	imdbRating string `json:"imdbRating"`
	Plot       string `json:"Plot"`
	SeriesID   string `json:"seriesID"`
}

type GenreMovie struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Genre      string `json:"Genre"`
	imdbRating string `json:"imdbRating"`
}

func getEpisode(w http.ResponseWriter, r *http.Request) {
	seriesTitle := r.URL.Query().Get("series_title")
	season := r.URL.Query().Get("season")
	episode := r.URL.Query().Get("episode_number")

	if seriesTitle == "" || season == "" || episode == "" {
		http.Error(w, "Missing parameters, please include series_title, season, episode_number", http.StatusBadRequest)
		return
	}
	url := "http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + url.QueryEscape(seriesTitle) + "&season=" + season + "&episode=" + episode
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to get episode", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Invalid response from omdb", http.StatusInternalServerError)
		return
	}
	fmt.Println(string(body))
	fmt.Println(url)

	var episodeResp EpisodeResponse
	if err := json.Unmarshal(body, &episodeResp); err != nil {
		fmt.Println("Error unmarshalling episode data:", err)
		fmt.Println(string(body))
		http.Error(w, "Invalid episode data from omdb", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(episodeResp)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	resp, e := http.Get("http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + title)
	if e != nil {
		http.Error(w, "Failed to fetch movie", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var movie MovieResponse
	if err := json.Unmarshal(body, &movie); err != nil {
		http.Error(w, "Failed to parse movie data", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movie)
}

func getMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	genre := r.URL.Query().Get("genre")
	if genre == "" {
		http.Error(w, "Missing genre parameter", http.StatusBadRequest)
		return
	}
	titles := []string{
		// Action
		"The Dark Knight",
		"Mad Max: Fury Road",
		"Gladiator",
		"Die Hard",
		"Casino Royale",

		// Sci-Fi
		"Blade Runner 2049",
		"The Matrix",
		"Inception",
		"Interstellar",
		"Arrival",

		// Drama
		"The Shawshank Redemption",
		"The Godfather",
		"Pulp Fiction",
		"Forrest Gump",
		"Parasite",

		// Comedy
		"Superbad",
		"Airplane!",
		"Shaun of the Dead",
		"The Grand Budapest Hotel",

		// Fantasy
		"The Lord of the Rings: The Fellowship of the Ring",
		"Pan's Labyrinth",
		"Harry Potter and the Sorcerer's Stone",

		// Animation
		"Spirited Away",
		"Toy Story",
		"Spider-Man: Into the Spider-Verse",
		"The Lion King",

		// Horror & Thriller
		"The Silence of the Lambs",
		"Get Out",
		"The Shining",
		"Seven",
	}

	var genreMovies []GenreMovie
	for i := 0; i < len(titles); i++ {
		title := titles[i]
		resp, err := http.Get("http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + url.QueryEscape(title))
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		var m GenreMovie
		if err := json.Unmarshal(body, &m); err != nil {
			continue
		}
		if m.Genre != "" && containsGenre(m.Genre, genre) {
			genreMovies = append(genreMovies, m)
		}
	}

	for i := 0; i < len(genreMovies)-1; i++ {
		for j := 0; j < len(genreMovies)-i-1; j++ {
			if parseRating(genreMovies[j].imdbRating) < parseRating(genreMovies[j+1].imdbRating) {
				temp := genreMovies[j]
				genreMovies[j] = genreMovies[j+1]
				genreMovies[j+1] = temp
			}
		}
	}

	if len(genreMovies) > 15 {
		genreMovies = genreMovies[:15]
	}

	json.NewEncoder(w).Encode(genreMovies)
}

func containsGenre(genres string, genre string) bool {
	genresLower := strings.ToLower(genres)
	genreLower := strings.ToLower(genre)
	if strings.Index(genresLower, genreLower) != -1 {
		return true
	}
	return false
}

func parseRating(rating string) float64 {
	val, err := strconv.ParseFloat(rating, 64)
	if err != nil {
		return 0.0
	}
	return val
}

func main() {
	godotenv.Load()
	apiKey = os.Getenv("OMDB_KEY")
	http.HandleFunc("/", webServerTest)
	http.HandleFunc("/api/movie", getMovie)
	http.HandleFunc("/api/episode", getEpisode)
	http.HandleFunc("/api/movies/genre", getMoviesByGenre)
	fmt.Println("starting server at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}
