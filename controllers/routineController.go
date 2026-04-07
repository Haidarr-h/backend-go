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

func GetRoutine(c *gin.Context) {

}

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
