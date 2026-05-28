// cmd/seed/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Haidarr-h/backend-go/internal/config"
	"github.com/Haidarr-h/backend-go/internal/exercise"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

const githubBaseURL = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/"

type sourceExercise struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Force            string   `json:"force"`
	Level            string   `json:"level"`
	Mechanic         string   `json:"mechanic"`
	Equipment        string   `json:"equipment"`
	PrimaryMuscles   []string `json:"primaryMuscles"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Instructions     []string `json:"instructions"`
	Category         string   `json:"category"`
	Images           []string `json:"images"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env")
	}

	db := config.InitDB()

	if err := db.AutoMigrate(&exercise.Exercise{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	data, err := os.ReadFile("/home/haidarhanif/Work/projects/group-projects/free-exercise-db/dist/exercises.json")
	if err != nil {
		log.Fatalf("failed to read exercises.json: %v", err)
	}

	var sources []sourceExercise
	if err := json.Unmarshal(data, &sources); err != nil {
		log.Fatalf("failed to parse exercises.json: %v", err)
	}

	var exercises []exercise.Exercise
	for _, s := range sources {
		images := make([]string, len(s.Images))
		for i, img := range s.Images {
			images[i] = githubBaseURL + img
		}

		exercises = append(exercises, exercise.Exercise{
			Name:             s.Name,
			Force:            s.Force,
			Level:            s.Level,
			Mechanic:         s.Mechanic,
			Equipment:        s.Equipment,
			PrimaryMuscles:   pq.StringArray(s.PrimaryMuscles),
			SecondaryMuscles: pq.StringArray(s.SecondaryMuscles),
			Instructions:     pq.StringArray(s.Instructions),
			Images:           pq.StringArray(images),
			Category:         s.Category,
		})
	}

	result := db.CreateInBatches(exercises, 100)
	if result.Error != nil {
		log.Fatalf("failed to seed: %v", result.Error)
	}

	fmt.Printf("seeded %d exercises\n", result.RowsAffected)
}