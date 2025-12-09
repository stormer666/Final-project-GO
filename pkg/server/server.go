package server

import (
	"final-project-go/pkg/api"
	"log"
	"net/http"
	"os"
)

// Run запускает веб-сервер
func Run() {
	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		port = envPort
	}

	api.Init()

	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Printf("Сервер запущен на http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
