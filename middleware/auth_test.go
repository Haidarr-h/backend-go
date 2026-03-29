package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Haidarr-h/backend-go/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// dummy test
func TestOnePlusOne(t *testing.T) {
	result := 1 + 1
	if result != 2 {
		t.Errorf("expected 2, got %d", result)
	}
}

// === helper functions ===
func setupRouter() *gin.Engine {
	// this makes the terminal not produced usual logs (since its for testing only)
	gin.SetMode(gin.TestMode)

	// bare router
	r := gin.New()
	r.Use(middleware.RequireAuth())

	// dummy route to test if the middleware works
	// remember our goal here is to test the middleware only, not the database, controller, and so on
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	return r

}

// helper: builds a signed JWT for testing
func makeTestToken(secret string, expired bool) string {
	expiry := time.Now().Add(1 * time.Hour)

	if expired {
		expiry = time.Now().Add(-1 * time.Hour)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": "user-123",
		"exp": expiry.Unix(),
	})

	signed, _ := token.SignedString([]byte(secret))
	return signed
}

// === test cases ===
func TestRequireAuth_NoHeader(t *testing.T) {
	os.Setenv("SECRET", "mysecret")
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	// notice: no header set

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization header")
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	os.Setenv("SECRET", "testsecret")
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.fake")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}
