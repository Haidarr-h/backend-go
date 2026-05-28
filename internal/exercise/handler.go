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
		exercise.GET("/{id}", h.GetExercise)
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

	exercises, err := h.service.GetExercises()

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
