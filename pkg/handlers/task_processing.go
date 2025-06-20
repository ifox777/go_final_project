package handlers

import (
	"errors"
	"go-final/pkg/scheduler"
	"time"
)

func ValidateAndProcessTaskRequest(req *TaskRequest, now time.Time) (time.Time, error) {
	if req.Title == "" {
		return time.Time{}, errors.New("Заголовок задачи не может быть пустым")
	}

	var finalDate time.Time

	if req.Date == "" || req.Date == "today" || req.Date == now.Format("20060102") {
		finalDate = now
	} else {
		parsedDate, err := time.ParseInLocation("20060102", req.Date, time.Local)
		if err != nil {
			return time.Time{}, errors.New("Invalid date format")
		}
		finalDate = parsedDate

		if finalDate.Before(now) {
			if req.Repeat == "" {
				finalDate = now
			} else {
				next, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat)
				if err != nil {
					return time.Time{}, err
				}
				finalDate, _ = time.ParseInLocation("20060102", next, time.Local)
			}
		}
	}

	if req.Repeat != "" && req.Date != "today" && req.Date != now.Format("20060102") {
		if _, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat); err != nil {
			return time.Time{}, err
		}
	}

	return finalDate, nil
}
