package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-final/pkg/scheduler"
	"net/http"
	"time"
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
		w.Header().Set("Content-Type",
			"application/json; charset=UTF-8")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Method Not Allowed"})
			if err != nil {
				return
			}

			return
		}

		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
			if err != nil {
				return
			}

			return
		}

		//валидация заголовка
		if req.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Title is required"})
			if err != nil {
				return
			}
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
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid date format"})
			if err != nil {
				return
			}
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
					err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
					if err != nil {
						return
					}
					return
				}
				finalDate, _ = time.Parse("20060102", next)
			}
		}

		// Проверка правила повторения
		if req.Repeat != "" {
			if _, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
				if err != nil {
					return
				}
				return
			}
		}

		//Добавление задачи в БД
		res, err := db.Exec(`INSERT INTO tasks (date, title, comment, repeat) 
VALUES (?, ?, ?, ?)`,
			finalDate.Format("20060102"),
			req.Title,
			req.Comment,
			req.Repeat,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			if err != nil {
				return
			}
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			if err != nil {
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(SuccessResponse{ID: id})
		if err != nil {
			return
		}
		_, err = fmt.Fprint(w, `{"id":`, id, "}")
		if err != nil {
			return
		}

	}

}
