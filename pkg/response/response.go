package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"sucess"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// 200 sections
func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// 400 sections
func BadRequest(c *gin.Context, message string, error any) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Error:   error,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
	})
}

func NotFound(c *gin.Context, message string, error any) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
		Error:   error,
	})
}

// 500 sections
func InternalError(c *gin.Context, message string, error any) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
		Error:   error,
	})
}
