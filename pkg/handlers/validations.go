package handlers

//
//
//func validateTask(req taskRequest)
//
//	if req.Title == "" {
//		respondWithError(w, http.StatusBadRequest, "Заголовок задачи не может быть пустым")
//		return
//	}
//
//	// Устанавливаем now как начало текущего дня (без времени)
//	now := time.Now().Local().Truncate(24 * time.Hour)
//	var finalDate time.Time
//
//	// Парсим дату или используем today
//	if req.Date == "" || req.Date == "today" || req.Date == now.Format("20060102") {
//		finalDate = now
//	} else {
//		parsedDate, err := time.ParseInLocation("20060102", req.Date, time.Local)
//		if err != nil {
//			respondWithError(w, http.StatusBadRequest, "Invalid date format")
//			return
//		}
//		parsedDate = parsedDate
//		finalDate = parsedDate
//
//		// Коррекция только для дат в прошлом (сравниваем как даты без времени)
//		if finalDate.Before(now) {
//			if req.Repeat == "" {
//				finalDate = now
//			} else {
//				next, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat)
//				if err != nil {
//					respondWithError(w, http.StatusBadRequest, err.Error())
//					return
//				}
//				finalDate, _ = time.ParseInLocation("20060102", next, time.Local)
//			}
//		}
//	}
//
//	// Валидация правила повтора (только если дата не today/now)
//	if req.Repeat != "" && req.Date != "today" && req.Date != now.Format("20060102") {
//		if _, err := scheduler.NextDate(now, finalDate.Format("20060102"), req.Repeat); err != nil {
//			respondWithError(w, http.StatusBadRequest, err.Error())
//			return
//		}
//	}
//
//}
