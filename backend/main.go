// backend/main.go
package main

import (
	"fmt"
	"log"

	"github.com/StepOne-ai/GoStresser/config"
	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/handlers"
	"github.com/StepOne-ai/GoStresser/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

    db.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    } else {
		log.Println("Connected to database")
	}

    db.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

    if err := db.AutoMigrate(db.DB); err != nil {
        log.Fatal("Failed to migrate tables:", err)
    }

    r := gin.Default()
    api := r.Group("/api/v1")
    {
        auth := api.Group("/auth")
        {
            auth.POST("/register", handlers.RegisterHandler)
            auth.POST("/login", handlers.LoginHandler)
        }

        scenario := api.Group("/scenarios", middleware.AuthMiddleware())
        {
            scenario.POST("/", handlers.CreateScenarioHandler)
            scenario.GET("/", handlers.GetScenariosHandler)
            scenario.GET("/:id", handlers.GetScenarioHandler)
            scenario.PUT("/:id", handlers.UpdateScenarioHandler)
            scenario.DELETE("/:id", handlers.DeleteScenarioHandler)
        }

        run := api.Group("/run", middleware.AuthMiddleware())
        {
            run.POST("/:id", handlers.StartTestHandler)
            run.GET("/:run_id/metrics", handlers.GetLiveMetricsHandler)
            run.GET("/:run_id/report", handlers.GetReportHandler)
        }
    }

    r.Run(":8080")
}