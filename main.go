package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"todo_restapi/internal/config"
	"todo_restapi/internal/http-server/handlers"
	"todo_restapi/internal/http-server/middlewares"
	"todo_restapi/internal/storage"
)

func main() {

	cfg := config.LoadConfig()

	database, err := storage.OpenStorage(cfg.StoragePath)
	if err != nil {
		log.Fatalf("OpenStorage: %v", err)
	}

	defer func() {
		if err := database.CloseStorage(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	taskHandler := handlers.NewTaskHandler(database, cfg)
	autService := middlewares.NewAuthService(cfg)

	router := chi.NewRouter()

	router.Get("/", func(write http.ResponseWriter, request *http.Request) {
		http.ServeFile(write, request, "web/index.html")
	})
	router.Handle("/*", http.StripPrefix("/", http.FileServer(http.Dir("web"))))

	router.Get("/api/nextdate", taskHandler.NextDate)

	router.Post("/api/signin", taskHandler.Authentication)
	router.With(middlewares.Auth(autService)).Route("/api", func(router chi.Router) {

		router.Get("/task", taskHandler.GetTask)
		router.Post("/task", taskHandler.AddTask)
		router.Put("/task", taskHandler.EditTask)
		router.Delete("/task", taskHandler.DeleteTask)

		router.Get("/tasks", taskHandler.GetTasks)
		router.HandleFunc("/task/done", taskHandler.TaskIsDone)
	})

	fmt.Printf("Server is running on port%s...\n", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, router); err != nil {
		log.Fatalf("server run error: %v\n", err)
	}
}
