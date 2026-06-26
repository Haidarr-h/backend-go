package bodymetrics

import (
	"errors"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type BodyMetricHandler struct {
	service Service
}

func NewBodyMetricHandler(service Service) *BodyMetricHandler {
	return &BodyMetricHandler{service: service}
}

func (h *BodyMetricHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	metrics := rg.Group("/body-metrics")
	{
		metrics.POST("/", h.UpsertBodyMetric)
		metrics.GET("/", h.GetBodyMetric)
	}
}

// UpsertBodyMetric godoc
// @Summary      Save my body metrics
// @Description  Create or overwrite the authenticated user's body metrics. BMI is derived from height and weight.
// @Tags         body-metrics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      UpsertBodyMetricRequest  true  "Body metrics"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/body-metrics [post]
func (h *BodyMetricHandler) UpsertBodyMetric(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	var req UpsertBodyMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Bad Request", err.Error())
		return
	}

	result, err := h.service.UpsertBodyMetric(userID, req)
	if err != nil {
		response.InternalError(c, "Failed to save body metrics", err.Error())
		return
	}

	response.OK(c, "Body metrics saved", result)
}

// GetBodyMetric godoc
// @Summary      Get my body metrics
// @Description  Get the authenticated user's body metrics (with derived BMI)
// @Tags         body-metrics
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/body-metrics [get]
func (h *BodyMetricHandler) GetBodyMetric(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.service.GetBodyMetric(userID)
	if err != nil {
		if errors.Is(err, ErrBodyMetricNotFound) {
			response.NotFound(c, "body metrics not found", err.Error())
			return
		}
		response.InternalError(c, "Failed to get body metrics", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}
