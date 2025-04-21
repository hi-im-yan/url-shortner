package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"url-shortner/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	r.Get("/short", s.redirectUrlHandler)
	r.Post("/short", s.shortLinkHandler)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}

func (s *Server) redirectUrlHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("Redirecting user")

	http.Redirect(w, r, "https://google.com", http.StatusSeeOther)
}

func (s *Server) shortLinkHandler(w http.ResponseWriter, r *http.Request) {

	// log.Printf("[routes:shortLinkHandler] rquest: %+v", r.Host)

	fmt.Printf("%s %s", r.URL.Scheme, r.Host)


	var reqBody struct {
		LinkToShort string `json:"link_to_short"`
		ExpTimeMinutes int `json:"exp_time_minutes"`
	}

	json.NewDecoder(r.Body).Decode(&reqBody)
	log.Printf("[routes:shortLinkHandler] Request received with body: %+v", reqBody)

	new := &database.ShortUrlModel{
		Link:           reqBody.LinkToShort,
		ExpTimeMinutes: reqBody.ExpTimeMinutes,
		ShortCode:      generateRandomString(8),
	}

	entity, err := s.db.SaveShortUrl(new)

	if err != nil {
		errResponse := struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}{
			Status:  500,
			Message: "Something went wrong with generating short url. Try again later",
		}

		json.NewEncoder(w).Encode(errResponse)
	}

	baseUrl := "http://"
	if r.URL.Scheme != "" {
		baseUrl = "https://"
	}

	succResponse := struct {
		Status int `json:"status"`
		ShortUrl string `json:"short_url"`
	} {
		Status: 200,
		ShortUrl: fmt.Sprint(baseUrl+r.Host+"/short/"+entity.ShortCode),
	}

	json.NewEncoder(w).Encode(succResponse)
}

func generateRandomString(stringLength int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	finalStringRune := make([]rune, stringLength)
	for i := range finalStringRune {
		finalStringRune[i] = letters[rand.Intn(len(letters))]
	}
	return string(finalStringRune)
}
