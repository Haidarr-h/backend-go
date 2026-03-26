package controllers

import (
	"net/http"

	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/gin-gonic/gin"
)

type ExerciseRequest struct {
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
	Equipment   string `json:"equipment"`
	Category    string `json:"category"`
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

	err := initializers.DB.Find(&exercises).Error
	// 1. Look at the type of exercises (to which table)
	// 2. fetch all rows from it
	// 3. Then populate exercises with the result

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": exercises,
	})

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": exercise,
	})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed to read request body",
			"detail": err.Error(),
		})
		return
	}

	// 2. Check if the same exercise already exist (return error if not found)
	var existingExercise models.Exercise
	err := initializers.DB.Where("name = ?", exerciseBody.Name).First(&existingExercise).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name already exist",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create exercise",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"messages": "exercise created successfully",
	})

}
