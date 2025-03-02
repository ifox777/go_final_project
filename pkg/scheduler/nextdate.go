package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату выполнения задачи
func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Парсим исходную дату
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("Ошибка парсинга исходной даты: %v", err)
	}

	// Обрабатываем правило повторения
	switch {
	case strings.HasPrefix(repeat, "d"): // Если повторяем каждые n дней
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("Неверный формат повторения: ожидается 'd n'")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", errors.New("Неверный формат повторения: ожидается 'd n', где n - число от 1 до 400")
		}

		// Добавляем количество дней, пока они не превысили now
		for !parsedDate.After(now) {
			parsedDate = parsedDate.AddDate(0, 0, days)
		}

	case repeat == "y": // Если повторяем ежегодно
		// Добавляем год, пока дата не превысила now
		for !parsedDate.After(now) {
			parsedDate = parsedDate.AddDate(1, 0, 0)
		}

	case strings.HasPrefix(repeat, "w"): // Если повторяем каждые n недель
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("Неверный формат повторения: ожидается 'w n'")
		}

		daysOfWeek, err := parseDaysOfWeek(parts[1])
		if err != nil {
			return "", err
		}

		// Добавляем недели, пока они не превысили now
		for {
			parsedDate = parsedDate.AddDate(0, 0, 1)
			if parsedDate.After(now) && containsDayOfWeek(daysOfWeek, parsedDate.Weekday()) {
				break
			}
		}

	case strings.HasPrefix(repeat, "m"): // Если повторяем каждый n месяц
		parts := strings.Split(repeat, " ")
		if len(parts) < 1 {
			return "", errors.New("Неверный формат повторения: ожидается 'm n'")
		}

		daysOfMonth, months, err := parseDaysAndMonths(parts[1:])
		if err != nil {
			return "", err
		}

		// Добавляем месяцы, пока они не превысили now
		for {
			parsedDate = parsedDate.AddDate(0, 0, 1)
			if parsedDate.After(now) && containsDayOfMonth(daysOfMonth, parsedDate.Year(),  parsedDate.Month(),  parsedDate.Day()) && containsMonth(months, parsedDate.Month()) {
				break
			}
		}

	default:
		return "", errors.New("Неверный формат повторения: ожидается 'd n' или 'y' или 'w n' или 'm n' или 'm n, n'")
	}

	return parsedDate.Format("20060102"), nil
}

// parseDaysOfWeek преобразует строку дней недели в массив Weekday
func parseDaysOfWeek(input string) ([]time.Weekday, error) {
	days := strings.Split(input, ",")
	result := make([]time.Weekday, 0, len(days))

	for _, day := range days {
		day = strings.TrimSpace(day) // Удаление пробелов из строки
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < 1 || dayInt > 7 {
			return nil, errors.New("Неверный формат повторения: ожидается 'w n', где n - число от 1 до 7")
		}
		result = append(result, time.Weekday(dayInt-1))
	}

	return result, nil
}

// parseDaysAndMonths преобразует строку дней и месяцев в массивы int
func parseDaysAndMonths(parts []string) ([]int, []time.Month, error) {
	if len(parts) == 0 {
		return nil, nil, errors.New("Неверный формат повторения: ожидается 'm n'")
	}

	days := strings.Split(parts[0], ",")
	daysOfMonth := make([]int, 0, len(days))

	for _, day := range days {
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < -2 || dayInt > 31 || dayInt == 0 {
			return nil, nil, errors.New("Неверный формат повторения: ожидается 'm n', где n - число от -2 до 31")
		}
		daysOfMonth = append(daysOfMonth, dayInt)
	}

	months := make([]time.Month, 0)
	if len(parts) > 1 {
		monthParts := strings.Split(parts[1], ",")
		for _, month := range monthParts {
			monthInt, err := strconv.Atoi(month)
			if err != nil || monthInt < 1 || monthInt > 12 {
				return nil, nil, errors.New("invalid month value")
			}
			months = append(months, time.Month(monthInt))
		}
	}

	return daysOfMonth, months, nil
}

// containsDayOfWeek проверяет вхождение дней в слайс
func containsDayOfWeek(days []time.Weekday, day time.Weekday) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}

// containsDayOfMonth проверяет облюдений правил для дня месяца
func containsDayOfMonth(days []int, targetYear int, targetMonth time.Month, targetDay int) bool {
	for _, day := range days {
		if day == -1 {
			lastDay := time.Date(targetYear, targetMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
			if targetDay == lastDay {
				return true
			}
		} else if day == -2 {
			lastDay := time.Date(targetYear, targetMonth+1, 0, 0, 0, 0, 0, time.UTC).Day()
			if targetDay == lastDay-1 {
				return true
			}
		} else if day == targetDay {
			return true
		}
	}

	return false
}

// containsMonth проверяет вхождение месяцев в слайс
func containsMonth(months []time.Month, month time.Month) bool {
	if len(months) == 0 {
		return true
	}

	for _, m := range months {
		if m == month {
			return true
		}
	}
	return false
}
