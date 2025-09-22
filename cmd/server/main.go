package main

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/you/tasks/internal/config"
	"github.com/you/tasks/internal/db"
	"github.com/you/tasks/internal/handler"
	"github.com/you/tasks/internal/model"
	"github.com/you/tasks/internal/repository"
	"github.com/you/tasks/internal/server"
)

func main() {
	// Загружаем конфиг
	cfg := config.Load()

	// Подключение к БД
	gormDB, err := db.ConnectWithRetry(cfg.DatabaseDSN, 12, 3*time.Second)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	// Миграция
	if err := gormDB.AutoMigrate(&model.Task{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Репозиторий и обработчики
	taskRepo := repository.NewTaskRepository(gormDB)
	taskHandler := handler.NewTaskHandler(taskRepo)

	// Echo и сервер
	e := echo.New()
	srv := server.NewServer(e)
	srv.RegisterRoutes(taskHandler)

	if err := srv.Start(cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
