package controllers

import (
	"errors"

	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExerciseRequest struct {
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
	Equipment   string `json:"equipment"`
	Category    string `json:"category"`
}

type ExerciseUpdateRequest struct {
	Id          int    `json:"id" binding:"required"`
	Name        string `json:"name,omitempty"`
	MuscleGroup string `json:"muscleGroup,omitempty"`
	Equipment   string `json:"equipment,omitempty"`
	Category    string `json:"category,omitempty"`
}

type ExerciseDelReq struct {
	Id []int `json:"id" binding:"required,min=1"`
}

// GetExercises godoc
// @Summary      Get all exercises
// @Tags         exercises
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/exercises [get]
func GetExercises(c *gin.Context) {
	var exercises []models.Exercise

	// check if there are any exercises
	if result := initializers.DB.First(&exercises); result.RowsAffected == 0 {
		response.OK(c, "No Exercise Found", "")
		return
	}

	err := initializers.DB.Find(&exercises).Error
	// 1. Look at the type of exercises (to which table)
	// 2. fetch all rows from it
	// 3. Then populate exercises with the result

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "No Exercises Found", err.Error())
			return
		}

		response.InternalError(c, "Error Fetching Exercise Data", err.Error())
		return
	}

	response.OK(c, "Successfully fetch all exercises", exercises)

}

// GetExercise godoc
// @Summary      Get a single exercise
// @Tags         exercises
// @Produce      json
// @Param        id   path      string  true  "Exercise ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/exercises/{id} [get]
func GetExercise(c *gin.Context) {
	exerciseId := c.Param("id")

	var exercise models.Exercise

	err := initializers.DB.First(&exercise, exerciseId).Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "No Exercises Found", err.Error())
			return
		}

		response.InternalError(c, "Error Fetching Exercise Data", err.Error())
		return
	}

	response.OK(c, "Successfully fetch exercise data", exercise)
}

// CreateExercise godoc
// @Summary      Create a new exercise
// @Tags         exercises
// @Accept       json
// @Produce      json
// @Param        body  body      ExerciseRequest  true  "Exercise data"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/exercises [post]
func CreateExercise(c *gin.Context) {
	var exerciseBody ExerciseRequest

	// 1. Read the requst body
	if err := c.ShouldBindJSON(&exerciseBody); err != nil {
		response.BadRequest(c, "Failed to read request body", err.Error())
		return
	}

	// 2. Check if the same exercise already exist (return error if not found)
	var existingExercise models.Exercise
	err := initializers.DB.Where("name = ?", exerciseBody.Name).First(&existingExercise).Error

	if err == nil {
		response.BadRequest(c, "Name already exist", "Bad Request")
		return
	}

	// 3. Send to database
	newExercise := models.Exercise{
		Name:        exerciseBody.Name,
		MuscleGroup: exerciseBody.MuscleGroup,
		Equipment:   exerciseBody.Equipment,
		Category:    exerciseBody.Category,
	}

	result := initializers.DB.Create(&newExercise)

	if result.Error != nil {
		response.InternalError(c, "Failed to create exercise", err.Error())
		return
	}

	response.Created(c, "exercise created successfully", newExercise)

}

// UpdateExercises godoc
// @Summary      Update a exercise
// @Tags         exercises
// @Accept       json
// @Produce      json
// @Param        body  body      ExerciseUpdateRequest  true  "Exercise data"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/exercises [patch]
func UpdateExercises(c *gin.Context) {
	var updateExercise ExerciseUpdateRequest

	// 1. Read Request body
	if err := c.ShouldBindJSON(&updateExercise); err != nil {
		response.BadRequest(c, "Wrong request Body", err.Error())
		return
	}

	// 2. builds the selected fields only
	updates := map[string]any{}
	if updateExercise.Name != "" {
		updates["name"] = updateExercise.Name
	}
	if updateExercise.MuscleGroup != "" {
		updates["muscleGroup"] = updateExercise.MuscleGroup
	}
	if updateExercise.Equipment != "" {
		updates["equipment"] = updateExercise.Equipment
	}
	if updateExercise.Category != "" {
		updates["category"] = updateExercise.Category
	}

	if len(updates) == 0 {
		response.BadRequest(c, "No fields to update", "")
		return
	}

	// 3. Updates the database
	err := initializers.DB.Model(models.Exercise{}).Where("id = ?", updateExercise.Id).Updates(updates).Error

	if err != nil {
		response.InternalError(c, "Failed to update user", err.Error())
		return
	}

	response.OK(c, "Update exercise successfully", "")
}

// DeleteExercises godoc
// @Summary      Delete exercises
// @Tags         exercises
// @Accept       json
// @Produce      json
// @Param        body  body      ExerciseDelReq  true  "Exercise data"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Router       /api/v1/exercises [delete]
func DeleteExercises(c *gin.Context) {
	var deleteExercise ExerciseDelReq
	var exercise models.Exercise

	// 1. binding json
	if err := c.ShouldBindJSON(&deleteExercise); err != nil {
		response.BadRequest(c, "Wrong request Body", err.Error())
		return
	}

	// 2. Delete
	result := initializers.DB.Delete(&exercise, deleteExercise.Id)

	if result.Error != nil {
		response.InternalError(c, "Failed deleting user", result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		response.NotFound(c, "No exercises found with given IDs", "")
		return
	}

	response.OK(c, "Successfully deleted exercises", "")
}
