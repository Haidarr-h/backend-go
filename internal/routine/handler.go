package routine

import (
	"errors"
	"strconv"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type RoutineHandler struct {
	routineService Service
}

func NewRoutineHandler(routineService Service) *RoutineHandler {
	return &RoutineHandler{routineService: routineService}
}

func (rc *RoutineHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	routine := rg.Group("/routines")
	{
		routine.POST("/", rc.CreateRoutine)
		routine.GET("/", rc.GetRoutines)
		routine.PATCH("/", rc.UpdateRoutine)
		routine.DELETE("/", rc.DeleteRoutine)
	}

}

// CreateRoutine godoc
// @Summary      Create Routine
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines [post]
func (rc *RoutineHandler) CreateRoutine(c *gin.Context) {

	// 1. get the usesr id from token
	userID := c.GetUint("userID")

	var routineCreateRequest CreateRoutineRequest

	// 2. bind the request
	if err := c.ShouldBindJSON(&routineCreateRequest); err != nil {
		response.BadRequest(c, "Bad Request", err.Error())
		return
	}

	result, err := rc.routineService.CreateRoutine(userID, routineCreateRequest)
	if err != nil {

		if errors.Is(err, ErrExerciseNotFound) {
			response.BadRequest(c, "Exercise not found", ErrExerciseNotFound)
			return
		}

		response.InternalError(c, "Internal error", err.Error())
		return
	}

	response.Created(c, "Routine Created", result)
}

// GetRoutines godoc
// @Summary      Get routines
// @Description  Get routines owned by the user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines [get]
func (rc *RoutineHandler) GetRoutines(c *gin.Context) {
	// 1. get user id from token
	userID := c.GetUint("userID")

	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	// 2. run the service
	result, err := rc.routineService.GetRoutines(userID)
	if err != nil {
		response.InternalError(c, "Internal Error: Couldn't Get Routines ID", err.Error())
		return
	}

	// 3. return the response
	response.OK(c, "Fetch Successfully", result)
}

// UpdateRoutine godoc
// @Summary      Update routine
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines/{id} [patch]
func (rc *RoutineHandler) UpdateRoutine(c *gin.Context) {
	// 1. get user id from token
	userID := c.GetUint("userID")

	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	// 2. get param
	routineId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid routine id", err.Error())
		return
	}

	// 3. bind the request
	var req UpdateRoutineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Body parsing failed", err.Error())
		return
	}

	// 4. call the service routine
	result, err := rc.routineService.UpdateRoutine(uint(routineId), req)
	if err != nil {
		response.InternalError(c, "Failed to update the user routine", err.Error())
		return
	}

	response.OK(c, "Success Updating user routine", result)
}

// DeleteRoutine godoc
// @Summary      delete routine
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines/{id} [delete]
func (rc *RoutineHandler) DeleteRoutine(c *gin.Context) {
	// 1. get user id from token
	userId := c.GetUint("userID")

	if userId == 0 {
		response.Unauthorized(c, "Unauthorized By Controller, Invalid Token")
		return
	}

	// 2. Get routine id from param
	routineId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Failed to get routine id from param", err.Error())
		return
	}

	// 3. call the service routine
	errRepo := rc.routineService.DeleteRoutine(uint(routineId), userId)
	if errRepo != nil {
		response.InternalError(c, "Failed to delete routien", errRepo.Error())
		return
	}

	response.OK(c, "Delete routine successful", "Success")
}
