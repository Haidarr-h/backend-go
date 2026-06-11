package health

import (
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Check server status
// @Description Perform status check to the server
// @Router /healthCheck [get]
func HealthCheck(c *gin.Context) {
	response.OK(c, "health status is ok", "server is running")
}
