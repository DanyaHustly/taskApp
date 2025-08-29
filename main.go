package main

import (
	"encoding/json"
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

// Task — модель GORM + JSON теги
type Task struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Task      string         `json:"task"`
	IsDone    bool           `json:"is_done"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

var db *gorm.DB

func main() {
	// Читаем DSN из окружения (docker-compose задаст DATABASE_DSN)
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		// значение по умолчанию для локальной разработки
		dsn = "postgres://myuser:mypass@localhost:5432/tasksdb?sslmode=disable"
	}

	var err error
	db, err = connectWithRetry(dsn, 12, 3*time.Second)
	if err != nil {
		panic("failed to connect db: " + err.Error())
	}

	// Миграция схемы (создаст таблицу tasks если её нет)
	if err := db.AutoMigrate(&Task{}); err != nil {
		panic("auto migrate failed: " + err.Error())
	}

	e := echo.New()

	// Роуты
	e.POST("/tasks", postTask)         // Create (сохранить в DB)
	e.GET("/tasks", listTasks)         // Read (все из DB)
	e.GET("/tasks/:id", getTask)       // Read (одна)
	e.PATCH("/tasks/:id", updateTask)  // Update (частичное)
	e.DELETE("/tasks/:id", deleteTask) // Delete (soft)

	e.Logger.Fatal(e.Start(":8080"))
}

// Попытки подключения с retry (полезно при docker-compose, чтобы ждать postgres)
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

// ===== POST /tasks ====
// Принимаем JSON { "task": "...", "is_done": true/false } и сохраняем в БД
func postTask(c echo.Context) error {
	var req struct {
		Task   string `json:"task"`
		IsDone bool   `json:"is_done"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
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

// ===== GET /tasks ====
// Возвращаем все таски (soft-deleted не выдаются автоматически GORM)
func listTasks(c echo.Context) error {
	var tasks []Task
	if err := db.Find(&tasks).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusOK, tasks)
}

// ===== GET /tasks/:id ====
// Вернуть одну задачу по id
func getTask(c echo.Context) error {
	id := c.Param("id")
	var t Task
	if err := db.First(&t, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}
	return c.JSON(http.StatusOK, t)
}

// ===== PATCH /tasks/:id ====
// Частичное обновление: клиент может прислать {"task":"...", "is_done":true}
// Используем map чтобы поддержать частичное обновление
func updateTask(c echo.Context) error {
	id := c.Param("id")

	var existing Task
	if err := db.First(&existing, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}

	var data map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&data); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	// Защита от изменения служебных полей
	delete(data, "id")
	delete(data, "created_at")
	delete(data, "updated_at")
	delete(data, "deleted_at")

	if err := db.Model(&existing).Updates(data).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}

	// Вернём обновлённую запись
	if err := db.First(&existing, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusOK, existing)
}

// ===== DELETE /tasks/:id ====
// Мягкое удаление: GORM пометит DeletedAt, запись не будет возвращаться в Find/First
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
