package user

import (
	"errors"

	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service Service
}

func NewUserHandler(service Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUserRoutes(rg *gin.RouterGroup) {
	user := rg.Group("/users")
	{
		user.GET("/me", h.GetMyData)
		user.DELETE("/", h.DeleteUser)
	}
}

// GetMyData godoc
// @Summary      Get users data (his data)
// @Tags         users/me
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetMyData(c *gin.Context) {
	
	userID := c.GetUint("userID")

	userData, err := h.service.GetMyData(userID)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.BadRequest(c, "user not found", ErrUserNotFound)
			return
		}
		response.InternalError(c, "internal error", err.Error())
		return
	}

	response.OK(c, "get user data successful", userData)
}

// DeleteUser godoc
// @Summary      delete a user
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 1. get id from param
	// userID, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	// if idErr != nil {
	// 	response.BadRequest(c, "invalid routine id", idErr.Error())
	// 	return
	// }

	// get from token instead
	userID := c.GetUint("userID")

	if err := h.service.Delete(uint(userID)); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.BadRequest(c, "user not found", ErrUserNotFound)
			return
		}

		response.InternalError(c, "internal error", err.Error())
		return
	}

	response.OK(c, "Delete user successful", "succes")
}


// // GetUsers godoc
// // @Summary      Get all users
// // @Tags         users
// // @Produce      json
// // @Success      200  {object}  map[string]interface{}
// // @Failure      500  {object}  map[string]interface{}
// // @Router       /api/v1/users [get]
// func GetUsers(c *gin.Context, cfg *config.Config) {
// 	var users []User

// 	// get all records
// 	result := cfg.DB.Find(&users)

// 	if result.Error != nil {
// 		response.InternalError(c, "Failed to fetch users data", "")
// 		return
// 	}

// 	response.OK(c, "Fetch users data success", users)
// }

// // GetUsers godoc
// // @Summary      Get a user
// // @Tags         users
// // @Produce      json
// // @Success      200  {object}  map[string]interface{}
// // @Failure      500  {object}  map[string]interface{}
// // @Router       /api/v1/users/{id} [get]
// func GetUser(c *gin.Context, cfg *config.Config) {
// 	// 1. Get param data (user id)
// 	userId := c.Param("id")

// 	var user User

// 	if err := cfg.DB.First(&user, userId).Error; err != nil {
// 		response.InternalError(c, "Failed to fetch user data", err.Error())
// 		return
// 	}

// 	response.OK(c, "Successfully fetch user data", user)

// }

// // UpdateUser godoc
// // @Summary      Update a user
// // @Tags         users
// // @Produce      json
// // @Param        body  body      updateReq  true  "user data"
// // @Success      200  {object}  map[string]interface{}
// // @Failure      500  {object}  map[string]interface{}
// // @Router       /api/v1/users/{id} [patch]
// func UpdateUser(c *gin.Context, cfg *config.Config) {
// 	var req updateReq

// 	// 1. Parse the body json
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.BadRequest(c, "Failed to parse body request", err.Error())
// 		return
// 	}

// 	// 2. Builds the selected fields map
// 	updates := map[string]any{}

// 	if req.FullName != "" {
// 		updates["fullName"] = req.FullName
// 	}

// 	if req.Username != "" {
// 		updates["usename"] = req.Username
// 	}

// 	// 3. update to database
// 	if err := cfg.DB.Model(User{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
// 		response.InternalError(c, "Failed to update user data to database", err.Error())
// 		return
// 	}

// 	response.OK(c, "User update successfully", "")
// }

