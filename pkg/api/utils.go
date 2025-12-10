package api

import (
	"encoding/json"
	"net/http"

	db "final-project-go/pkg/database"
)

const dateFormat = "20060102"

type TaskResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type TaskRequest struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// TasksResponse структура для ответа со списком задач
type TasksResponse struct {
	Tasks  []db.Task `json:"tasks"`
	Result string    `json:"error,omitempty"`
}

// writeJSON записывает JSON ответ
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeJSONError записывает JSON ответ с ошибкой
func writeJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	response := TaskResponse{
		Error: errorMsg,
	}
	writeJSON(w, response, statusCode)
}

// writeJSONSuccess записывает пустой успешный JSON ответ
func writeJSONSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}
