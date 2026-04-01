package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	db "final-project-go/pkg/database"
)

// DoneHandler обрабатывает POST запросы для отметки задачи выполненной
func DoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if task.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {

		if task.Date == "" {
			writeJSONError(w, "task date is empty", http.StatusBadRequest)
			return
		}

		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Failed to calculate next date: %v", err), http.StatusBadRequest)
			return
		}

		err = db.UpdateDate(id, nextDate)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	writeJSONSuccess(w)
}
