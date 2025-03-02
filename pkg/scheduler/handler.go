package scheduler

import (
	"fmt"
	"net/http"
	"time"
)

// NexDateHandler обрабатывает запросы к  /api/nextdate
func NexDateHandler(w http.ResponseWriter, r *http.Request) {
		// Получаем парметры из запроса
		nowStr := r.URL.Query().Get("now")
		dateStr := r.URL.Query().Get("date")
		repeatRule := r.URL.Query().Get("repeat")

		// Паррсим текущще время
		now, err := parseTime(nowStr)
		if err != nil {
			http.Error(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}

		//Вычисляем следующую дату
		NextDate, err := NextDate(now, dateStr, repeatRule)
		if err != nil {
			http.Error(w, "Неверный формат даты", http.StatusBadRequest)
			return
		}

		//Возвращаем результат
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w,NextDate)
}


// parseTime парсит время
func parseTime(nowStr string) (time.Time, error) {
	timeParse, err  := time.Parse("2006-01-02T15:04:05Z", nowStr)

	return timeParse, err
}
