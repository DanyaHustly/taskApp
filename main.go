package main

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Task struct {
	ID     string `json:"id"`
	Task   string `json:"task"`
	IsDone bool   `json:"is_done"`
}

var tasks = []Task{}

func postTask(c echo.Context) error {
	var req Task
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	req.ID = uuid.New().String()
	tasks = append(tasks, req)

	return c.JSON(http.StatusCreated, req)
}

func listTasks(c echo.Context) error {
	return c.JSON(http.StatusOK, tasks)
}

func getTask(c echo.Context) error {
	id := c.Param("id")
	for _, t := range tasks {
		if t.ID == id {
			return c.JSON(http.StatusOK, t)
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
}

func updateTask(c echo.Context) error {
	id := c.Param("id")

	var req Task
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	for i, t := range tasks {
		if t.ID == id {
			if req.Task != "" {
				tasks[i].Task = req.Task
			}
			tasks[i].IsDone = req.IsDone
			return c.JSON(http.StatusOK, tasks[i])
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
}

func deleteTask(c echo.Context) error {
	id := c.Param("id")
	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return c.NoContent(http.StatusNoContent)
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"error": "task not found"})
}
func main() {
	e := echo.New()

	e.POST("/tasks", postTask)        // C
	e.GET("/tasks", listTasks)        // R (all)
	e.GET("/tasks/:id", getTask)      // R (one)
	e.PATCH("/tasks/:id", updateTask) // U (частично)
	e.DELETE("/tasks/:id", deleteTask)

	e.Logger.Fatal(e.Start(":8080"))

}
