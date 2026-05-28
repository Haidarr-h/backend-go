package exercise

import (
	"strconv"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type ExerciseHandler struct {
	service Service
}

func NewExerciseHandler(service Service) *ExerciseHandler {
	return &ExerciseHandler{service: service}
}

func (h *ExerciseHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	exercise := rg.Group("/exercises")
	{
		exercise.GET("/", h.GetExercises)
		exercise.GET("/:id", h.GetExercise)
	}
}

// GetExercises godoc
// @Summary      Get all exercises
// @Tags         exercises
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/exercises [get]
func (h *ExerciseHandler) GetExercises(c *gin.Context) {

	category := c.Query("category")
	equipment := c.Query("equipment")
	level := c.Query("level")
	mechanic := c.Query("equipmmechanicent")
	muscleGroup := c.Query("muscleGroup")
	limit := c.DefaultQuery("limit", "0")
	offset := c.DefaultQuery("offset", "0")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		response.BadRequest(c, "invalid limit value", "invalid limit")
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		response.BadRequest(c, "invalid offset value", "invalid offset")
		return
	}

	filterParams := ExerciseFilterParams{
		Category:    category,
		MuscleGroup: muscleGroup,
		Level:       level,
		Equipment:   equipment,
		Mechanic:    mechanic,
		Limit:       limitInt,
		Offset:      offsetInt,
	}

	exercises, err := h.service.GetExercises(filterParams)

	if err != nil {
		response.InternalError(c, "Failed to get all exercises", err.Error())
		return
	}

	response.OK(c, "successfully get all exercises data", exercises)
}

// GetExercise godoc
// @Summary      Get a single exercise
// @Tags         exercises
// @Produce      json
// @Param        id   path      string  true  "Exercise ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/exercises/{id} [get]
func (h *ExerciseHandler) GetExercise(c *gin.Context) {

	exerciseId, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid exercise id", idErr.Error())
		return
	}

	req := ExerciseRequest{
		ID: uint(exerciseId),
	}

	exercise, err := h.service.GetExercise(req)

	if err != nil {
		response.InternalError(c, "Failed to get exercise data", err.Error())
		return
	}

	response.OK(c, "successfully get all exercises data", exercise)
}
