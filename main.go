package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"hello-chi/dbShit"
	_ "hello-chi/docs"
	"hello-chi/handlers"
	"hello-chi/utils"
	"log"
	"net/http"
	"os"
)

func main() {
	utils.Init()
	r := chi.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	rdb := dbShit.InitializeRedisClient()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	r.Post("/upload/initiate", handlers.InitiateUploadClosure(rdb))
	r.Put("/upload/{identifier}", handlers.UploadChunkHandlerClosure(rdb))
	r.Get("/download/{identifier}", handlers.DownloadHandlerClosure(rdb))
	r.Post("/upload/internal", handlers.UploadFileInternalClosure(rdb))
	r.Post("/zip", handlers.DownloadAsZipHandlerClosure(rdb))
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	log.Println("Starting server on :" + os.Getenv("PORT") + "..................")
	port := os.Getenv("PORT")

	handler := c.Handler(r)

	err := http.ListenAndServe(":"+port, handler)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

}
