package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ===== Модель для БД =====
type Task struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Task      string         `json:"task"`
	IsDone    bool           `json:"is_done"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ===== DTO (входные данные) =====
type CreateTaskRequest struct {
	Task   string `json:"task" validate:"required"`
	IsDone bool   `json:"is_done"`
}

type UpdateTaskRequest struct {
	Task   *string `json:"task"`
	IsDone *bool   `json:"is_done"`
}

var db *gorm.DB

func main() {
	// Читаем DSN из окружения
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://myuser:mypass@localhost:5432/tasksdb?sslmode=disable"
	}

	var err error
	db, err = connectWithRetry(dsn, 12, 3*time.Second)
	if err != nil {
		panic("failed to connect db: " + err.Error())
	}

	// Авто-миграция
	if err := db.AutoMigrate(&Task{}); err != nil {
		panic("auto migrate failed: " + err.Error())
	}

	e := echo.New()

	// Роуты
	e.POST("/tasks", postTask)
	e.GET("/tasks", listTasks)
	e.GET("/tasks/:id", getTask)
	e.PATCH("/tasks/:id", updateTask)
	e.DELETE("/tasks/:id", deleteTask)

	e.Logger.Fatal(e.Start(":8080"))
}

// ===== Retry подключение к Postgres =====
func connectWithRetry(dsn string, attempts int, wait time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 0; i < attempts; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			sqlDB, err2 := db.DB()
			if err2 == nil {
				if err3 := sqlDB.Ping(); err3 == nil {
					return db, nil
				}
			}
		}
		fmt.Printf("DB connect attempt %d/%d failed: %v. retrying in %s...\n", i+1, attempts, err, wait)
		time.Sleep(wait)
	}
	return nil, err
}

// ===== POST /tasks =====
func postTask(c echo.Context) error {
	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	if req.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "task is required"})
	}

	t := Task{
		ID:     uuid.New().String(),
		Task:   req.Task,
		IsDone: req.IsDone,
	}
	if err := db.Create(&t).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusCreated, t)
}

// ===== GET /tasks =====
func listTasks(c echo.Context) error {
	var tasks []Task
	if err := db.Find(&tasks).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusOK, tasks)
}

// ===== GET /tasks/:id =====
func getTask(c echo.Context) error {
	id := c.Param("id")
	var t Task
	if err := db.First(&t, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}
	return c.JSON(http.StatusOK, t)
}

// ===== PATCH /tasks/:id =====
func updateTask(c echo.Context) error {
	id := c.Param("id")

	var existing Task
	if err := db.First(&existing, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	// Обновляем только то, что пришло
	if req.Task != nil {
		existing.Task = *req.Task
	}
	if req.IsDone != nil {
		existing.IsDone = *req.IsDone
	}

	if err := db.Save(&existing).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}

	return c.JSON(http.StatusOK, existing)
}

// ===== DELETE /tasks/:id =====
func deleteTask(c echo.Context) error {
	id := c.Param("id")
	var t Task
	if err := db.First(&t, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}
	if err := db.Delete(&t).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.NoContent(http.StatusNoContent)
}
