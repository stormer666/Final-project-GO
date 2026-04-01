package api

import (
	"fmt"
	"net/http"

	db "final-project-go/pkg/database"
)

const maxTasksLimit = 50

// TasksHandler обрабатывает GET запросы для получения списка задач
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := db.Tasks(maxTasksLimit)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Failed to get tasks: %v", err), http.StatusInternalServerError)
		return
	}

	response := TasksResponse{
		Tasks: tasks,
	}

	if response.Tasks == nil {
		response.Tasks = []db.Task{}
	}

	writeJSON(w, response, http.StatusOK)
}
