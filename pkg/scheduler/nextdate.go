package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет слудующую дату выполнения задачи
func NextDate(now time.Time, date string, repeat string) (string, error) {

	//Парсим исходную дату
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("Ошибка парсинга исходной даты: %v", err)
	}
	//Обрабатываем правло повторения
	switch {
	case strings.HasPrefix(repeat, "d"): //Если повторяем каждые n дней
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return "", errors.New("Неверный формат повторения: ожидается 'd n'")

		}

		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", errors.New("Неверный формат повторения: ожидается 'd n', где n - число от 1 до 400")
		}

		//Добавляем количество дней, пока они не превысили now
		for !parsedDate.After(now) {
			parsedDate = parsedDate.AddDate(0, 0, days)
		}

	}
}
