package main

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"

	"taskApp/internal/db"
	"taskApp/internal/handler"
	"taskApp/internal/model"
	"taskApp/internal/repository"
)

func main() {
	// Читаем DSN из окружения
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://myuser:mypass@localhost:5432/tasksdb?sslmode=disable"
	}

	// Подключаемся к БД (используем internal/db.ConnectWithRetry)
	gormDB, err := db.ConnectWithRetry(dsn, 12, 3*time.Second)
	if err != nil {
		panic("failed to connect db: " + err.Error())
	}

	// Авто-миграция (модель перенесена в internal/model)
	if err := gormDB.AutoMigrate(&model.Task{}); err != nil {
		panic("auto migrate failed: " + err.Error())
	}

	// Репозиторий
	taskRepo := repository.NewTaskRepository(gormDB)

	// Хендлеры (используют репозиторий)
	taskHandler := handler.NewTaskHandler(taskRepo)

	e := echo.New()

	// Роуты (те же, что были в оригинале)
	e.POST("/tasks", taskHandler.PostTask)
	e.GET("/tasks", taskHandler.ListTasks)
	e.GET("/tasks/:id", taskHandler.GetTask)
	e.PATCH("/tasks/:id", taskHandler.UpdateTask)
	e.DELETE("/tasks/:id", taskHandler.DeleteTask)

	e.Logger.Fatal(e.Start(":8080"))
}
