package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
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
func BadRequest(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: message,
	})
}

func Forbidden(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}

func Conflict(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}

func Gone(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusGone, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}

func NotFound(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}

// 500 sections
func InternalError(c *gin.Context, message string, errorMes any) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
		Error:   errorMes,
	})
}
