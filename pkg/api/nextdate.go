package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// NextDate вычисляет следующую дату выполнения задачи
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("repeat rule is empty")
	}

	startDate, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	ruleType, args, err := parseRepeatRule(repeat)
	if err != nil {
		return "", err
	}

	switch ruleType {
	case "d":
		return nextDateByDays(now, startDate, args)
	case "y":
		return nextDateYearly(now, startDate)
	default:
		return "", fmt.Errorf("unknown repeat type: %s", ruleType)
	}
}

// parseRepeatRule разбирает правило повторения
func parseRepeatRule(repeat string) (string, []string, error) {
	if len(repeat) == 0 {
		return "", nil, fmt.Errorf("repeat rule is empty")
	}

	parts := make([]string, 0)
	start := 0
	inWord := false

	for i, ch := range repeat {
		if ch == ' ' {
			if inWord {
				parts = append(parts, repeat[start:i])
				inWord = false
			}
		} else {
			if !inWord {
				start = i
				inWord = true
			}
		}
	}

	if inWord {
		parts = append(parts, repeat[start:])
	}

	if len(parts) == 0 {
		return "", nil, fmt.Errorf("invalid repeat format")
	}

	return parts[0], parts[1:], nil
}

// nextDateByDays обрабатывает правило "d <число>"
func nextDateByDays(now, startDate time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("invalid d format, expected 'd <days>'")
	}

	days, err := parseInt(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid days number: %w", err)
	}

	if days <= 0 || days > 400 {
		return "", fmt.Errorf("days must be between 1 and 400")
	}

	date := startDate

	for {
		date = date.AddDate(0, 0, days)
		if afterNow(date, now) {
			break
		}
	}

	return date.Format(dateFormat), nil
}

// nextDateYearly обрабатывает правило "y"
func nextDateYearly(now, startDate time.Time) (string, error) {
	date := startDate

	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}

	return date.Format(dateFormat), nil
}

// parseInt преобразует строку в целое число
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// afterNow проверяет, что date > now (только дата, без времени)
func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return date.After(now)
}

// NextDateHandler обрабатывает запросы к /api/nextdate
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	if dateStr == "" || repeatStr == "" {
		http.Error(w, "Parameters 'date' and 'repeat' are required", http.StatusBadRequest)
		return
	}

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid 'now' parameter: %v", err), http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(nextDate)); err != nil {
		log.Printf("Ошибка при записи ответа: %v", err)
	}
}
