package controllers

import (
	"github.com/Haidarr-h/backend-go/config"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type updateReq struct {
	Id       int    `json:"id" binding:"required,min=1"`
	FullName string `json:"fullName"`
	Username string `json:"username"`
}

// GetUsers godoc
// @Summary      Get all users
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users [get]
func GetUsers(c *gin.Context, cfg *config.Config) {
	var users []models.User

	// get all records
	result := cfg.DB.Find(&users)

	if result.Error != nil {
		response.InternalError(c, "Failed to fetch users data", "")
		return
	}

	response.OK(c, "Fetch users data success", users)
}

// GetUsers godoc
// @Summary      Get a user
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/{id} [get]
func GetUser(c *gin.Context, cfg *config.Config) {
	// 1. Get param data (user id)
	userId := c.Param("id")

	var user models.User

	if err := cfg.DB.First(&user, userId).Error; err != nil {
		response.InternalError(c, "Failed to fetch user data", err.Error())
		return
	}

	response.OK(c, "Successfully fetch user data", user)

}

// UpdateUser godoc
// @Summary      Update a user
// @Tags         users
// @Produce      json
// @Param        body  body      updateReq  true  "user data"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/{id} [patch]
func UpdateUser(c *gin.Context, cfg *config.Config) {
	var req updateReq

	// 1. Parse the body json
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Failed to parse body request", err.Error())
		return
	}

	// 2. Builds the selected fields map
	updates := map[string]any{}

	if req.FullName != "" {
		updates["fullName"] = req.FullName
	}

	if req.Username != "" {
		updates["usename"] = req.Username
	}

	// 3. update to database
	if err := cfg.DB.Model(models.User{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		response.InternalError(c, "Failed to update user data to database", err.Error())
		return
	}

	response.OK(c, "User update successfully", "")
}

// UpdateUser godoc
// @Summary      delete a user
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/{id} [delete]
func DeleteUser(c *gin.Context, cfg *config.Config) {
	// 1. get id from param
	id := c.Param("id")

	user := models.User{}

	// 2. check if user exist
	if err := cfg.DB.First(&user, id).Error; err != nil {
		response.BadRequest(c, "User not found", err.Error())
		return
	}

	// 3. Delete all associated records + user in one transaction
	txErr := cfg.DB.Transaction(func(tx *gorm.DB) error {
		// routine_exercises must go first (FK → routines)
		routineIDs := tx.Model(&models.Routine{}).Select("id").Where("user_id = ?", id)
		if err := tx.Unscoped().Where("routine_id IN (?)", routineIDs).Delete(&models.RoutineExercises{}).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.Routine{}).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_id = ?", id).Delete(&models.OtpVerification{}).Error; err != nil {
			return err
		}

		return tx.Delete(&user).Error
	})

	if txErr != nil {
		response.InternalError(c, "Failed to delete user", txErr.Error())
		return
	}

	response.OK(c, "Delete user successful", user)
}
