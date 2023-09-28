package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Location struct {
	Name string `json:"name"`
}

type Origin struct {
	Name string `json:"name"`
}

type Character struct {
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Species  string   `json:"species"`
	Gender   string   `json:"gender"`
	Image    string   `json:"image"`
	Url      string   `json:"url"`
	Created  string   `json:"created"`
	Location Location `json:"location"`
	Episode  []string `json:"episode"`
}

func removeDuplicates(characters []Character) []Character {
	uniqueNames := make(map[string]bool)

	var uniqueCharacters []Character

	for _, char := range characters {
		if !uniqueNames[char.Name] {
			uniqueNames[char.Name] = true
			uniqueCharacters = append(uniqueCharacters, char)
		}
	}

	return uniqueCharacters
}

func main() {
	log.Println("Starting server on port 3000...")

	jsonFile, err := os.Open("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var characters []Character
	json.Unmarshal(byteValue, &characters)

	uniqueCharacters := removeDuplicates(characters)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string][]Character{
			"Characters": uniqueCharacters,
		}

		templates := template.Must(template.ParseFiles("templates/index.html", "templates/characters.html"))
		if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.Get("/characters", func(w http.ResponseWriter, r *http.Request) {
		searchInput := r.URL.Query().Get("search_input")

		var results []Character

		if searchInput == "" {
			results = uniqueCharacters
		} else {
			for _, char := range uniqueCharacters {
				if strings.Contains(char.Name, searchInput) {
					results = append(results, char)
				}
			}
		}

		if len(results) == 0 {
			tmpl := template.Must(template.ParseFiles("templates/no_results.html"))
			tmpl.Execute(w, nil)
		} else {
			data := map[string][]Character{
				"Characters": results,
			}

			tmpl := template.Must(template.ParseFiles("templates/characters.html"))
			tmpl.Execute(w, data)
		}
	})

	log.Fatal(http.ListenAndServe(":3000", mux))
}
