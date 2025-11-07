package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type SuccessResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func Error(c *gin.Context, statusCode int, errorType, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error:     errorType,
		Message:   message,
		Timestamp: time.Now(),
	})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Message:   "created",
		Data:      data,
		Timestamp: time.Now(),
	})
}