package main

import (
	"database/sql"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/misgorod/hackaton/handler"
	"log"
	"net/http"
)

func main() {
	connStr := "user=postgres password=hackaton dbname=hackaton host=hackatondb.cgpygcvzbwp1.eu-central-1.rds.amazonaws.com port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	userHandler := handler.User{db}
	meetingHandler := handler.Meeting{db}
	healthHandler := handler.Health{}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
	r.Route("/api", func(r chi.Router) {
		r.Route("/participants", func(r chi.Router) {
			r.Post("/", userHandler.Post)
			r.Post("/{id}/meetings", meetingHandler.Post)
			//r.Get("/{id}/meetings", meetingHandler.GetAll)
			r.Put("/{ownerId}/meetings/{meetingId}", meetingHandler.Put)
		})

		r.Get("/healthcheck", healthHandler.Get)
	})
	log.Fatal(http.ListenAndServe(":80", r))
}
