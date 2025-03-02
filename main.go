package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	"go-final/pkg/database"
	"go-final/pkg/handlers"
	"go-final/pkg/scheduler"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultPort = "7540"
	webDir      = "./web"
)

var db *sql.DB

func main() {
	// Инициализация базы данных
	var err error
	db, err = database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Print("Файл .env не найден")
	}

	// Настройка порта
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}

	// Настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", scheduler.NextDateHandler)
	http.HandleFunc("/api/task", handlers.AddTaskHandler(db))

	// Запуск сервера
	log.Printf("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
