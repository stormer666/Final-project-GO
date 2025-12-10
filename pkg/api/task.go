package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "final-project-go/pkg/database"
)

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("TaskHandler: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut, http.MethodPatch:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// deleteTaskHandler обрабатывает DELETE запросы для удаления задач
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJSONError(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		writeJSONError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSONSuccess(w)
}

// getTaskHandler обрабатывает GET запросы для получения задачи по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJSONError(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		writeJSONError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	var err error

	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			writeJSONError(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			writeJSONError(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		req.Date = r.FormValue("date")
		req.Title = r.FormValue("title")
		req.Comment = r.FormValue("comment")
		req.Repeat = r.FormValue("repeat")
	}

	if req.Title == "" {
		writeJSONError(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := validateAndPrepareTask(req, false)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	response := TaskResponse{
		ID: fmt.Sprintf("%d", id),
	}

	writeJSON(w, response, http.StatusOK)
}

// updateTaskHandler обрабатывает PUT запросы для обновления задач
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	contentType := r.Header.Get("Content-Type")

	idFromURL := r.URL.Query().Get("id")

	var jsonRaw map[string]*json.RawMessage

	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&jsonRaw); err != nil {
			writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		rawBytes, _ := json.Marshal(jsonRaw)
		if err := json.Unmarshal(rawBytes, &req); err != nil {
			writeJSONError(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		if req.ID == "" && idFromURL != "" {
			req.ID = idFromURL
		}

	} else {
		if err := r.ParseForm(); err != nil {
			writeJSONError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		req.ID = r.FormValue("id")
		req.Date = r.FormValue("date")
		req.Title = r.FormValue("title")
		req.Comment = r.FormValue("comment")
		req.Repeat = r.FormValue("repeat")

		if req.ID == "" && idFromURL != "" {
			req.ID = idFromURL
		}
	}

	if req.ID == "" {
		writeJSONError(w, "task id is required", http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(req.ID); err != nil {
		writeJSONError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	existing, err := db.GetTask(req.ID)
	if err != nil {
		writeJSONError(w, "task not found", http.StatusNotFound)
		return
	}

	if req.Date != "" {
		existing.Date = req.Date
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Comment != "" {
		existing.Comment = req.Comment
	}

	if strings.Contains(contentType, "application/json") {
		if _, exists := jsonRaw["repeat"]; exists {
			existing.Repeat = req.Repeat
		}
	} else {
		if r.Form.Has("repeat") {
			existing.Repeat = req.Repeat
		}
	}

	if req.Repeat != "" {
		if req.Repeat != "y" && !strings.HasPrefix(req.Repeat, "d ") {
			writeJSONError(w, "invalid repeat format", http.StatusBadRequest)
			return
		}
	}

	if _, err := validateAndPrepareTask(TaskRequest{
		Date:    existing.Date,
		Title:   existing.Title,
		Comment: existing.Comment,
		Repeat:  existing.Repeat,
	}, true); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateTask(existing); err != nil {
		writeJSONError(w, fmt.Sprintf("failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSONSuccess(w)
}

// validateAndPrepareTask валидирует и подготавливает задачу
func validateAndPrepareTask(req TaskRequest, isUpdate bool) (*db.Task, error) {

	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	now := time.Now()

	if req.Date == "" {
		req.Date = now.Format(dateFormat)
	} else {
		if len(req.Date) != 8 {
			return nil, fmt.Errorf("invalid date format, expected YYYYMMDD")
		}

		for _, ch := range req.Date {
			if ch < '0' || ch > '9' {
				return nil, fmt.Errorf("invalid date format, expected YYYYMMDD")
			}
		}
	}

	taskDate, err := time.Parse(dateFormat, req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date")
	}

	year, month, day := taskDate.Date()
	if year < 2000 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return nil, fmt.Errorf("invalid date")
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	taskDate = time.Date(taskDate.Year(), taskDate.Month(), taskDate.Day(), 0, 0, 0, 0, taskDate.Location())

	if isUpdate {
		if taskDate.Before(now) {
			return nil, fmt.Errorf("date cannot be in the past")
		}
	}

	if req.Repeat != "" {
		if req.Repeat != "y" && !strings.HasPrefix(req.Repeat, "d ") {
			return nil, fmt.Errorf("invalid repeat format")
		}

		if strings.HasPrefix(req.Repeat, "d ") {
			parts := strings.Split(req.Repeat, " ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid repeat format")
			}

			days, err := strconv.Atoi(parts[1])
			if err != nil || days < 1 || days > 400 {
				return nil, fmt.Errorf("invalid repeat format")
			}
		}
	}

	if !isUpdate {
		if taskDate.Before(now) {
			if req.Repeat == "" {
				req.Date = now.Format(dateFormat)
			} else {
				nextDate, err := NextDate(now, req.Date, req.Repeat)
				if err != nil {
					return nil, fmt.Errorf("invalid repeat for past date: %v", err)
				}
				req.Date = nextDate
			}
		}
	}

	task := &db.Task{
		ID:      req.ID,
		Date:    req.Date,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	return task, nil
}
