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

var titles = []string{
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
	"Titanic",

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
	http.HandleFunc("/api/recommendation", getRecommendation)
	fmt.Println("starting server at http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}

type RecommendationMovie struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Genre      string `json:"Genre"`
	Director   string `json:"Director"`
	Actors     string `json:"Actors"`
	ImdbRating string `json:"imdbRating"`
}

func getRecommendation(w http.ResponseWriter, r *http.Request) {
	favorite := r.URL.Query().Get("favorite_movie")
	if favorite == "" {
		http.Error(w, "Missing favorite_movie parameter", http.StatusBadRequest)
		return
	}

	// Fetch favorite movie details
	resp, err := http.Get("http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + url.QueryEscape(favorite))
	if err != nil {
		http.Error(w, "Failed to fetch favorite movie", http.StatusInternalServerError)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read favorite movie response", http.StatusInternalServerError)
		return
	}
	var fav RecommendationMovie
	if err := json.Unmarshal(body, &fav); err != nil {
		http.Error(w, "Failed to parse favorite movie data", http.StatusInternalServerError)
		return
	}

	// Prepare lists for recommendations
	var genreRecs []RecommendationMovie
	var directorRecs []RecommendationMovie
	var actorRecs []RecommendationMovie

	for i := 0; i < len(titles); i++ {
		title := titles[i]
		if strings.ToLower(title) == strings.ToLower(fav.Title) {
			continue // skip the favorite movie itself
		}
		resp, err := http.Get("http://www.omdbapi.com/?apikey=" + apiKey + "&t=" + url.QueryEscape(title))
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		var m RecommendationMovie
		if err := json.Unmarshal(body, &m); err != nil {
			continue
		}
		if len(genreRecs) < 20 && fav.Genre != "" && m.Genre != "" && containsAny(fav.Genre, m.Genre) {
			genreRecs = append(genreRecs, m)
		}
		if len(directorRecs) < 20 && fav.Director != "" && m.Director != "" {
			if sharesDirector(fav.Director, m.Director) {
				directorRecs = append(directorRecs, m)
			}
		}
		if len(actorRecs) < 20 && fav.Actors != "" && m.Actors != "" {
			if sharesActor(fav.Actors, m.Actors) {
				actorRecs = append(actorRecs, m)
			}
		}
	}

	simpleSort(genreRecs)
	simpleSort(directorRecs)
	simpleSort(actorRecs)

	result := map[string]interface{}{
		"recommendation": map[string][]RecommendationMovie{
			"genre":    genreRecs,
			"director": directorRecs,
			"actor":    actorRecs,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func containsAny(genres1, genres2 string) bool {
	gs1 := strings.Split(genres1, ",")
	gs2 := strings.Split(genres2, ",")
	for i := 0; i < len(gs1); i++ {
		for j := 0; j < len(gs2); j++ {
			if strings.TrimSpace(strings.ToLower(gs1[i])) == strings.TrimSpace(strings.ToLower(gs2[j])) {
				return true
			}
		}
	}
	return false
}

func sharesActor(actors1, actors2 string) bool {
	as1 := strings.Split(actors1, ",")
	as2 := strings.Split(actors2, ",")
	for i := 0; i < len(as1); i++ {
		for j := 0; j < len(as2); j++ {
			if strings.TrimSpace(strings.ToLower(as1[i])) == strings.TrimSpace(strings.ToLower(as2[j])) {
				return true
			}
		}
	}
	return false
}
func sharesDirector(directors1, directors2 string) bool {
	d1 := strings.Split(directors1, ",")
	d2 := strings.Split(directors2, ",")

	for i := 0; i < len(d1); i++ {
		for j := 0; j < len(d2); j++ {
			if strings.TrimSpace(strings.ToLower(d1[i])) == strings.TrimSpace(strings.ToLower(d2[j])) {
				return true
			}
		}
	}
	return false
}

func simpleSort(movies []RecommendationMovie) {
	for i := 0; i < len(movies)-1; i++ {
		for j := 0; j < len(movies)-i-1; j++ {
			if parseRating(movies[j].ImdbRating) < parseRating(movies[j+1].ImdbRating) {
				temp := movies[j]
				movies[j] = movies[j+1]
				movies[j+1] = temp
			}
		}
	}
}
