package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var (
	// DB глобальное подключение к базе данных
	DB *sql.DB

	// schema SQL схема для создания таблицы
	schema = `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
		title VARCHAR(255) NOT NULL DEFAULT "",
		comment TEXT,
		repeat VARCHAR(128)
	);
	
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
)

// Task структура задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Init инициализирует соединение с базой данных
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Printf("Opening the DB failed: %v", err)
		return err
	}

	if install {
		_, err = db.Exec(schema)
		if err != nil {
			log.Printf("Create DB using a schema failed: %v", err)
			return err
		}
		log.Printf("Database created: %s", dbFile)
	}

	DB = db
	log.Printf("DB connected to %s", dbFile)
	return nil
}

// AddTask добавляет задачу в базу данных
func AddTask(date, title, comment, repeat string) (int64, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(query, date, title, comment, repeat)
	if err != nil {
		return 0, fmt.Errorf("failed to add task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetTask возвращает задачу по ID
func GetTask(id string) (*Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := DB.QueryRow(query, id)

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// GetAllTasks возвращает все задачи отсортированные по дате
func GetAllTasks() ([]Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

// UpdateTask обновляет задачу
func UpdateTask(task *Task) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	result, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// DeleteTask удаляет задачу
func DeleteTask(id string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	result, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func Tasks(limit int) ([]Task, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if tasks == nil {
		tasks = []Task{}
	}

	return tasks, nil
}

// UpdateDate обновляет дату задачи
func UpdateDate(id string, date string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	result, err := DB.Exec(query, date, id)
	if err != nil {
		return fmt.Errorf("failed to update date: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
