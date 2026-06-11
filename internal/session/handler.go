package session

import (
	"errors"
	"strconv"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService Service
}

func NewSessionHandler(sessionService Service) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	session := rg.Group("/sessions")
	{
		session.POST("/", h.CreateSession)
		session.GET("/", h.GetSessions)
		session.GET("/:id", h.GetSession)
		session.PATCH("/:id", h.UpdateSession)
		session.DELETE("/:id", h.DeleteSession)
	}
}

// CreateSession godoc
// @Summary      Create session
// @Description  Start a session from a routine; its exercises are snapshotted from the routine
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateSessionRequest  true  "Session to create"
// @Success      201   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Bad Request", err.Error())
		return
	}

	result, err := h.sessionService.CreateSession(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrRoutineNotFound):
			response.BadRequest(c, "routine not found", err.Error())
		case errors.Is(err, ErrRoutineNotAccessible):
			response.Forbidden(c, "this routine cannot be used", err.Error())
		case errors.Is(err, ErrInvalidTimeRange):
			response.BadRequest(c, "invalid time range", err.Error())
		default:
			response.InternalError(c, "Failed to create session", err.Error())
		}
		return
	}

	response.Created(c, "Session created", result)
}

// GetSessions godoc
// @Summary      Get sessions
// @Description  Get all sessions owned by the authenticated user
// @Tags         sessions
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/sessions [get]
func (h *SessionHandler) GetSessions(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	result, err := h.sessionService.GetSessions(userID)
	if err != nil {
		response.InternalError(c, "Failed to get sessions", err.Error())
		return
	}

	response.OK(c, "Fetch successfully", result)
}

// GetSession godoc
// @Summary      Get session
// @Description  Get a single session (with computed metrics) owned by the user
// @Tags         sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Session ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      403   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	sessionID, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid session id", idErr.Error())
		return
	}

	result, err := h.sessionService.GetSession(userID, uint(sessionID))
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			response.NotFound(c, "session not found", err.Error())
		case errors.Is(err, InvalidSessionOwnership):
			response.Forbidden(c, "you cannot access this session", err.Error())
		default:
			response.InternalError(c, "internal error", err.Error())
		}
		return
	}

	response.OK(c, "get session data successful", result)
}

// UpdateSession godoc
// @Summary      Update session
// @Description  Update a session's non-exercise fields (name, notes, times)
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int               true  "Session ID"
// @Param        body  body      UpdateSessionReq  true  "Fields to update"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      403   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/sessions/{id} [patch]
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	sessionID, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid session id", idErr.Error())
		return
	}

	var req UpdateSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Body parsing failed", err.Error())
		return
	}

	result, err := h.sessionService.UpdateSession(uint(sessionID), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			response.NotFound(c, "session not found", err.Error())
		case errors.Is(err, InvalidSessionOwnership):
			response.Forbidden(c, "you cannot access this session", err.Error())
		case errors.Is(err, ErrInvalidTimeRange):
			response.BadRequest(c, "invalid time range", err.Error())
		default:
			response.InternalError(c, "Failed to update session", err.Error())
		}
		return
	}

	response.OK(c, "Success updating session", result)
}

// DeleteSession godoc
// @Summary      Delete session
// @Description  Delete a session owned by the authenticated user
// @Tags         sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Session ID"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      401   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /api/v1/sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		response.Unauthorized(c, "Unauthorized by controller. Invalid token")
		return
	}

	sessionID, idErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if idErr != nil {
		response.BadRequest(c, "invalid session id", idErr.Error())
		return
	}

	if err := h.sessionService.DeleteSession(uint(sessionID), userID); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			response.NotFound(c, "session not found", err.Error())
			return
		}
		response.InternalError(c, "Failed to delete session", err.Error())
		return
	}

	response.OK(c, "Delete session successful", "success")
}
