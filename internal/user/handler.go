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
		user.DELETE("/me", h.DeleteUser)
	}
}

// GetMyData godoc
// @Summary      Get users data (his data)
// @Tags         users/me
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetMyData(c *gin.Context) {

	userID := c.GetUint("userID")

	userData, err := h.service.GetMyData(userID)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "user not found", err.Error())
			return
		}
		response.InternalError(c, "internal error", err.Error())
		return
	}

	response.OK(c, "get user data successful", userData)
}

// DeleteUser godoc
// @Summary      Delete the authenticated user
// @Tags         users/me
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/v1/users/me [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// the user being deleted is always the caller, taken from the JWT
	userID := c.GetUint("userID")

	if err := h.service.Delete(userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "user not found", err.Error())
			return
		}

		response.InternalError(c, "internal error", err.Error())
		return
	}

	response.OK(c, "Delete user successful", "success")
}

