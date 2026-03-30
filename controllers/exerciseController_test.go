package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haidarr-h/backend-go/controllers"
	"github.com/Haidarr-h/backend-go/initializers"
	"github.com/Haidarr-h/backend-go/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setup
func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect test db:", err)
	}
	db.AutoMigrate(&models.Exercise{})
	initializers.DB = db
}

// Test 1: success case
func TestGetExercises_Success(t *testing.T) {
	setupTestDB(t)

	// seed fake data in to the memory db
	initializers.DB.Create(&models.Exercise{
		Name:        "Bench Press",
		MuscleGroup: "Chest",
		Equipment:   "Barbell",
		Category:    "Compound",
	})

	initializers.DB.Create(&models.Exercise{
		Name:        "Squat",
		MuscleGroup: "Legs",
		Equipment:   "Barbell",
		Category:    "Compound",
	})

	// setup fake router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/exercises", controllers.GetExercises)

	// fire a fake GET request
	req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "Successfully fetch all exercises", body["message"])

	// make sure the body is not empty
	data := body["data"].([]interface{})
	assert.Equal(t, 2, len(data))
}

// Test 2: empty db
func TestGetExercises_Empty(t *testing.T) {
	setupTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/exercises", controllers.GetExercises)

	// send the requst
	req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, "No Exercise Found", body["message"])

	// when empty, data will be nil
	// data, ok := body["data"].([]interface{})
	// if !ok {
	// 	// data is null, which means empty — still valid
	// 	assert.Nil(t, body["data"])
	// 	return
	// }
	// assert.Equal(t, 0, len(data))
}
