package handler

import (
	"net/http"
	"taskApp/internal/model"

	"github.com/labstack/echo/v4"
	"taskApp/internal/repository"
)

// TaskHandler — использует репозиторий вместо прямого доступа к gorm.DB
type TaskHandler struct {
	repo repository.TaskRepository
}

func NewTaskHandler(r repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: r}
}

// DTOs — оставлены как в оригинале
type CreateTaskRequest struct {
	Task   string `json:"task" validate:"required"`
	IsDone bool   `json:"is_done"`
}

type UpdateTaskRequest struct {
	Task   *string `json:"task"`
	IsDone *bool   `json:"is_done"`
}

// PostTask — соответствует оригинальной реализации, только использует repo
func (h *TaskHandler) PostTask(c echo.Context) error {
	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	if req.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "task is required"})
	}

	t := &model.Task{
		Task:   req.Task,
		IsDone: req.IsDone,
	}
	if err := h.repo.Create(t); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *TaskHandler) ListTasks(c echo.Context) error {
	tasks, err := h.repo.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c echo.Context) error {
	id := c.Param("id")
	t, err := h.repo.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	if t == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}
	return c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) UpdateTask(c echo.Context) error {
	id := c.Param("id")
	t, err := h.repo.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	if t == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}

	if req.Task != nil {
		t.Task = *req.Task
	}
	if req.IsDone != nil {
		t.IsDone = *req.IsDone
	}

	if err := h.repo.Update(t); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) DeleteTask(c echo.Context) error {
	id := c.Param("id")
	t, err := h.repo.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	if t == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
	}
	if err := h.repo.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	return c.NoContent(http.StatusNoContent)
}
