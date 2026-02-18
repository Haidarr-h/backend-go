package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Check server status
// @Description Perform status check to the server
// @Router /healthCheck [get] 
func HealthCheck(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"Health Status": "ok",
		})
}