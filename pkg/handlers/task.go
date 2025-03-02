package handlers

import (
	"database/sql"
	"encoding/json"
	"go-final/pkg/scheduler"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type DBTask struct {
	ID      int    `db:"id"`
	Date    string `db:"date"`
	Title   string `db:"title"`
	Comment string `db:"comment"`
	Repeat  string `db:"repeat"`
}

type JSONTask struct {
	ID      string `json:"id"`
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

		// Валидация заголовка
		if req.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Title is required"})
			if err != nil {
				return
			}
			return
		}

		now := time.Now().UTC()
		var finalDate time.Time

		// Обработка даты
		if req.Date == "" || req.Date == "today" {
			finalDate = now // Устанавливаем текущую дату
		} else {
			parsedDate, err := time.Parse("20060102", req.Date)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid date format"})
				return
			}
			finalDate = parsedDate
		}

		log.Printf("Initial final date: %s, %s", finalDate.Format("20060102"), req.Title)

		// Коррекция даты только если не "today"
		if req.Date != "today" && finalDate.Before(now) {
			if req.Repeat == "" {
				finalDate = now
			} else {
				next, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
					return
				}
				finalDate, _ = time.Parse("20060102", next)
			}
		} else if req.Date == "today" {
			finalDate = now
		}

		log.Printf("Corrected final date: %s, %s", finalDate.Format("20060102"), req.Title)

		// Проверка правила повторения (только если дата не "today")
		if req.Repeat != "" && req.Date != "today" {
			if _, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
				return
			}
		} else if req.Date == "today" {
			finalDate = now
		}

		log.Printf("Final date after repeat check: %s, %s", finalDate.Format("20060102"), req.Title)

		// Добавление задачи в БД
		res, err := db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) 
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
	}
}

func GetTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		search := r.URL.Query().Get("search")
		var tasks []DBTask
		var err error
		ctx := r.Context()

		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE 1=1`
		args := []interface{}{}
		limit := 50

		if search != "" {
			if date, err := time.Parse("20060102", search); err == nil {
				//Поиск по дате
				query += " WHERE date = ?"
				args = append(args, date.Format("20060102"))
			} else {
				//ПОиск по подстройке
				query += " WHERE title LIKE ? OR comment LIKE ?"
				searchTerm := "%" + search + "%"
				args = append(args, searchTerm, searchTerm)
			}
		}

		query += " ORDER BY date LIMIT ?"
		args = append(args, limit)

		rows, err := db.QueryContext(ctx, query, args...)

		if err != nil {
			handleError(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		for rows.Next() {
			var task DBTask
			if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
				handleError(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, task)
		}

		if err := rows.Err(); err != nil {
			handleError(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Преобразование данных в JSONTask
		jsonTasks := make([]JSONTask, 0, len(tasks))
		for _, task := range tasks {
			jsonTasks = append(jsonTasks, JSONTask{
				ID:      strconv.Itoa(task.ID),
				Date:    task.Date,
				Title:   task.Title,
				Comment: task.Comment,
				Repeat:  task.Repeat,
			})
		}

		response := struct {
			Tasks []JSONTask `json:"tasks"`
		}{
			Tasks: jsonTasks,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			handleError(w, "Internal Server Error", http.StatusInternalServerError)
		}

	}
}

func handleError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(ErrorResponse{Error: message})
	if err != nil {
		return
	}
}

// GET /api/task
func GetTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Не указан ID задачи"})
			if err != nil {
				return
			}
			return
		}

		var task DBTask
		err := db.QueryRowContext(r.Context(), "SELECT id,"+
			" date, title, comment, repeat FROM scheduler WHERE id = ?",
			id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Задача не найдена"})
				if err != nil {
					return
				}
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal Server Error"})
			if err != nil {
				return
			}
			return
		}

		json.NewEncoder(w).Encode(JSONTask{
			ID:      strconv.Itoa(task.ID),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		})

	}
}

// PUT api/task
func UpdateTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
			if err != nil {
				return
			}
			return
		}

		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Не указан ID задачи"})
			if err != nil {
				return
			}
			return
		}

		//Валидация данных при обновлении задачи
		if req.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Не указан заголовок задачи"})
			if err != nil {
				return
			}
			return
		}

		now := time.Now().UTC()
		var finalDate time.Time

		// Обработка даты
		if req.Date != "" || req.Date == "today" {
			finalDate = now
		} else {
			parsedDate, err := time.Parse("20060102", req.Date)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Некорректная дата"})
				if err != nil {
					return
				}
				return
			}
			finalDate = parsedDate
		}

		// Проверка правила повторения (только если дата не "today")
		if req.Repeat != "" && req.Date != "today" {
			if _, err := scheduler.NextDate(now, finalDate.Format("20060102"),
				req.Repeat); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
				if err != nil {
					return
				}
				return
			}
		}

		// Обновление задачи в БД
		res, err := db.Exec(`UPDATE scheduler SET date = ?, title = ?, 
                     comment = ?, repeat = ? WHERE id = ?`,
			finalDate.Format("20060102"),
			req.Title,
			req.Comment,
			req.Repeat,
			req.ID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			if err != nil {
				return
			}
			return
		}

		rowsAffected, err := res.RowsAffected()
		if rowsAffected == 0 {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(ErrorResponse{Error: "Задача не найдена"})
			if err != nil {
				return
			}
			return
		}

		json.NewEncoder(w).Encode(struct{}{})

	}
}

// handlers/task.go
func MarkDoneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Не указан идентификатор"})
			return
		}

		var task struct {
			Date   string
			Repeat string
		}
		err := db.QueryRowContext(r.Context(),
			"SELECT date, repeat FROM scheduler WHERE id = ?", id).
			Scan(&task.Date, &task.Repeat)

		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Задача не найдена"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Ошибка сервера"})
			return
		}

		now := time.Now().UTC()
		var result sql.Result

		if task.Repeat != "" {
			next, err := scheduler.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
				return
			}
			result, err = db.ExecContext(r.Context(),
				"UPDATE scheduler SET date = ? WHERE id = ?", next, id)
		} else {
			result, err = db.ExecContext(r.Context(),
				"DELETE FROM scheduler WHERE id = ?", id)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Ошибка сервера"})
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Задача не найдена"})
			return
		}

		json.NewEncoder(w).Encode(struct{}{})
	}
}

func DeleteTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Не указан идентификатор"})
			return
		}

		result, err := db.ExecContext(r.Context(),
			"DELETE FROM scheduler WHERE id = ?", id)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Ошибка сервера"})
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Задача не найдена"})
			return
		}

		json.NewEncoder(w).Encode(struct{}{})
	}
}
