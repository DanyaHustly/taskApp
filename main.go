package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var task string

type requestBody struct {
	Task string `json:"task"`
}

func postTask(c echo.Context) error {
	var rb requestBody
	if err := c.Bind(&rb); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}

	task = rb.Task

	return c.JSON(http.StatusOK, map[string]string{
		"message": "task saved",
		"task":    task,
	})
}

func getTask(c echo.Context) error {
	return c.JSON(http.StatusOK, "Задача на сегодня: "+task)
}

func main() {
	e := echo.New()

	e.GET("/", getTask)
	e.POST("/task", postTask)

	e.Logger.Fatal(e.Start(":8080"))

}
