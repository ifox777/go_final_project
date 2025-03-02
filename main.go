package main

import (
	"fmt"
	"go-final/pkg/database"
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

func main() {

	if err := database.InitDB(); err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v\n", err)
	}
	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = defaultPort
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Некорректный порт: %s\n", port)
		return
	}
	// настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", scheduler.NexDateHandler)

	//Запуск сервера
	fmt.Printf("Сервер запущен на http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Ошибка при запуске сервера: %v\n", err)
	}

}
