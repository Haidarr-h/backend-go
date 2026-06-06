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
		routine.GET("/:id", rc.GetRoutine)
		routine.PATCH("/:id", rc.UpdateRoutine)
		routine.DELETE("/:id", rc.DeleteRoutine)
	}
}

// CreateRoutine godoc
// @Summary      Create Routine
// @Description  Create a new routine owned by the authenticated user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateRoutineRequest  true  "Routine to create"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
// @Description  Get all routines owned by the authenticated user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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

// GetRoutine godoc
// @Summary      Get routine
// @Description  Get a single routine owned by the user, or any public routine
// @Tags         routines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Routine ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/routines/{id} [get]
func (rc *RoutineHandler) GetRoutine(c *gin.Context) {
	// 1. get user id from token
	userID := c.GetUint("userID")

	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	// 2. get routine id
	routineID, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid routine id", idErr.Error())
		return
	}

	// 2. service
	routineData, err := rc.routineService.GetRoutine(userID, uint(routineID))

	if err != nil {
		if errors.Is(err, InvalidRoutineOwnership) {
			response.BadRequest(c, "user cannot access this routine", err)
			return
		}
		response.InternalError(c, "internal error", err)
		return
	}

	response.OK(c, "get routine data successful", routineData)
}

// UpdateRoutine godoc
// @Summary      Update routine
// @Description  Partially update a routine owned by the authenticated user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int               true  "Routine ID"
// @Param        body  body      UpdateRoutineReq  true  "Fields to update"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
	result, err := rc.routineService.UpdateRoutine(uint(routineId), userID, req)
	if err != nil {

		if errors.Is(err, InvalidRoutineOwnership) {
			response.BadRequest(c, "this routine cant be accessed by this account", InvalidRoutineOwnership)
			return
		}

		if errors.Is(err, ErrRoutineNotFound) {
			response.BadRequest(c, "routine not found", ErrRoutineNotFound)
			return
		}

		response.InternalError(c, "Failed to update the user routine", err.Error())
		return
	}

	response.OK(c, "Success Updating user routine", result)
}

// DeleteRoutine godoc
// @Summary      Delete routine
// @Description  Delete a routine owned by the authenticated user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Routine ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
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
