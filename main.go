package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/vitaemendum/rss/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	// 7.35.52
	fmt.Println("Hello, World!")

	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT environment variable not set")
	} else {
		fmt.Println(portString)
	}

	dbURL := os.Getenv("DB_URL")
	if portString == "" {
		log.Fatal("DB_URL environment variable not set")
	} else {
		fmt.Println(dbURL)
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cant connect to db :( ", err)
	}

	queries := database.New(conn)

	apiCfg := apiConfig{
		DB: queries,
	}

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/ready", handlerReadiness)
	v1Router.Get("/err", handlerErr)

	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

	v1Router.Get("/follows", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowsGet))
	v1Router.Post("/follows", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowCreate))
	v1Router.Delete("/follows/{id}", apiCfg.middlewareAuth(apiCfg.handlerFeedFollowDelete))

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf(":%s", portString),
	}

	log.Printf("Server starting on port %v", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
