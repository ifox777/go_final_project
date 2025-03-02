package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"go-final/pkg/scheduler"
)

type TaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	ID int64 `json:"id"`
}

func AddTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}

		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON format"})
			return
		}

		// Валидация обязательного заголовка
		if req.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Title is required"})
			return
		}

		// Обработка даты
		now := time.Now().UTC()
		if req.Date == "" {
			req.Date = now.Format("20060102")
		}

		// Парсинг даты
		parsedDate, err := time.Parse("20060102", req.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid date format"})
			return
		}

		// Коррекция даты
		finalDate := parsedDate
		if parsedDate.Before(now) {
			if req.Repeat == "" {
				finalDate = now
			} else {
				next, err := scheduler.NextDate(now, req.Date, req.Repeat)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
					return
				}
				finalDate, _ = time.Parse("20060102", next)
			}
		}

		// Проверка правила повторения
		if req.Repeat != "" {
			if _, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
				return
			}
		}

		// Вставка в базу данных
		res, err := db.Exec(
			`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
			finalDate.Format("20060102"),
			req.Title,
			req.Comment,
			req.Repeat,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get ID"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SuccessResponse{ID: id})
	}
}
