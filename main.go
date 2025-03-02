package main

import (
	"database/sql"
	"fmt"
	"go-final/pkg/database"
	"go-final/pkg/handlers"
	"go-final/pkg/scheduler"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const defaultPort = "7540"
const webDir = "./web"
const dbFile = "scheduler.db"

var db *sql.DB

func main() {
	var err error
	db, err = database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v\n", err)
	}
	defer db.Close()

	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Некорректный порт: %s\n", port)
		return
	}
	// настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", scheduler.NexDateHandler)
	http.HandleFunc("/api/task", handlers.AddTaskHandler(db))

	//Запуск сервера
	fmt.Printf("Сервер запущен на http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Ошибка при запуске сервера: %v\n", err)
	}

}
