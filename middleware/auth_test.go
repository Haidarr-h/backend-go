package middleware_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/Haidarr-h/backend-go/config"
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
func setupRouter(cfg *config.Config) *gin.Engine {
    // this makes the terminal not produced usual logs (since its for testing only)
    gin.SetMode(gin.TestMode)

    // bare router
    r := gin.New()
    r.Use(middleware.RequireAuth(cfg))  // Pass cfg here

    // dummy route to test if the middleware works
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

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{  // Changed to HS256
        "sub": "user-123",
        "exp": expiry.Unix(),
    })

    signed, _ := token.SignedString([]byte(secret))
    return signed
}

// === test cases ===
func TestRequireAuth_NoHeader(t *testing.T) {
    cfg := &config.Config{
        JWTSecret: "mysecret",
    }
    r := setupRouter(cfg)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    // notice: no header set

    r.ServeHTTP(w, req)

    assert.Equal(t, 401, w.Code)
    assert.Contains(t, w.Body.String(), "Missing authorization header")
}

func TestRequireAuth_InvalidToken(t *testing.T) {
    cfg := &config.Config{
        JWTSecret: "testsecret",
    }
    r := setupRouter(cfg)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer this.is.fake")
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
    assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// Additional test: Valid token should pass
func TestRequireAuth_ValidToken(t *testing.T) {
    secret := "testsecret"
    cfg := &config.Config{
        JWTSecret: secret,
    }
    r := setupRouter(cfg)

    validToken := makeTestToken(secret, false)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+validToken)
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "ok")
}