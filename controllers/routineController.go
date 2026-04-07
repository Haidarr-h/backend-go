package controllers

import (
	"github.com/Haidarr-h/backend-go/dto"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/Haidarr-h/backend-go/services"
	"github.com/gin-gonic/gin"
)

type RoutineController struct {
	routineService *services.RoutineService
}

func NewRoutineController(routineService *services.RoutineService) *RoutineController {
	return &RoutineController{routineService: routineService}
}

// CreateRoutine godoc
// @Summary      Create Routine
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines [post]
func (rc *RoutineController) CreateRoutine(c *gin.Context) {

	// 1. get the usesr id from token
	userID := c.GetUint("userID")

	var routineCreateRequest dto.CreateRoutineRequest

	// 2. bind the request
	if err := c.ShouldBindJSON(&routineCreateRequest); err != nil {
		response.BadRequest(c, "Bad Request", err.Error())
		return
	}

	result, err := rc.routineService.CreateRoutine(userID, routineCreateRequest)
	if err != nil {
		response.InternalError(c, "Internal error", err.Error())
		return
	}

	response.Created(c, "Routine Created", result)

}

// GetRoutines godoc
// @Summary      Get all routines owned by user
// @Tags         routines
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/routines [get]
func (rc *RoutineController) GetRoutines(c *gin.Context) {
	// 1. get user id from token
	userID := c.GetUint("userID")

	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
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

func (rc *RoutineController) UpdateRoutine(c *gin.Context) {
	// 
}
