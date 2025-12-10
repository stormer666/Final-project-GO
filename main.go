package main

import (
	db "final-project-go/pkg/database"
	"final-project-go/pkg/server"
	"log"
	"os"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	log.Printf("Инициализация базы данных: %s", dbFile)

	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}

	defer db.Close()

	log.Println("База данных готова")

	server.Run()
}
