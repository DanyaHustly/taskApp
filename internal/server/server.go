package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"taskApp/internal/handler"
)

type Server struct {
	e *echo.Echo
}

func NewServer(e *echo.Echo) *Server {
	return &Server{e: e}
}

func (s *Server) RegisterRoutes(taskHandler *handler.TaskHandler) {
	s.e.POST("/tasks", taskHandler.PostTask)
	s.e.GET("/tasks", taskHandler.ListTasks)
	s.e.GET("/tasks/:id", taskHandler.GetTask)
	s.e.PATCH("/tasks/:id", taskHandler.UpdateTask)
	s.e.DELETE("/tasks/:id", taskHandler.DeleteTask)
}

func (s *Server) Start(port string) error {
	addr := fmt.Sprintf(":%s", port)
	return s.e.Start(addr)
}
