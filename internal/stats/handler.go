package stats

import (
	"errors"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService Service
}

func NewStatsHandler(statsService Service) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	stats := rg.Group("/stats")
	{
		stats.GET("/career", h.GetCareerStats)
		stats.GET("/records", h.GetPersonalRecords)
		stats.GET("/volume", h.GetVolumeOverTime)
		stats.GET("/streak", h.GetStreak)
		stats.GET("/muscles", h.GetMuscleBreakdown)
	}
}

// GetCareerStats godoc
// @Summary      Get my career stats
// @Description  Lifetime totals (volume, reps, sets, sessions) plus level/XP/progress
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/stats/career [get]
func (h *StatsHandler) GetCareerStats(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.statsService.GetCareerStats(userID)
	if err != nil {
		response.InternalError(c, "Failed to get career stats", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}

// GetPersonalRecords godoc
// @Summary      Get my personal records
// @Description  Heaviest weight and best single-set volume per exercise
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/stats/records [get]
func (h *StatsHandler) GetPersonalRecords(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.statsService.GetPersonalRecords(userID)
	if err != nil {
		response.InternalError(c, "Failed to get personal records", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}

// GetVolumeOverTime godoc
// @Summary      Get my volume over time
// @Description  Volume, reps and session counts bucketed by week (default) or month
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Param        period  query     string  false  "Bucket size"  Enums(week, month)
// @Success      200     {object}  map[string]interface{}
// @Failure      400     {object}  map[string]interface{}
// @Failure      401     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /api/v1/stats/volume [get]
func (h *StatsHandler) GetVolumeOverTime(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.statsService.GetVolumeOverTime(userID, c.Query("period"))
	if err != nil {
		if errors.Is(err, ErrInvalidPeriod) {
			response.BadRequest(c, "invalid period", err.Error())
			return
		}
		response.InternalError(c, "Failed to get volume over time", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}

// GetStreak godoc
// @Summary      Get my workout streak
// @Description  Current and longest streak (up to 2 rest days between workouts keeps it alive)
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/stats/streak [get]
func (h *StatsHandler) GetStreak(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.statsService.GetStreak(userID)
	if err != nil {
		response.InternalError(c, "Failed to get streak", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}

// GetMuscleBreakdown godoc
// @Summary      Get my muscle volume breakdown
// @Description  Total volume per primary muscle across all sessions
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/stats/muscles [get]
func (h *StatsHandler) GetMuscleBreakdown(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.statsService.GetMuscleBreakdown(userID)
	if err != nil {
		response.InternalError(c, "Failed to get muscle breakdown", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}
