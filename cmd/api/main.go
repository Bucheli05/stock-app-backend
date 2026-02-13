package main

import (
	"context"
	"fmt"
	"github.com/Bucheli05/stock-app-backend/internal/config"
	"github.com/Bucheli05/stock-app-backend/internal/handlers"
	"github.com/Bucheli05/stock-app-backend/internal/repository"
	"github.com/Bucheli05/stock-app-backend/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := repository.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create repository and ensure table exists
	stockRepo := repository.NewStockRepository(db)
	ctx := context.Background()
	if err := stockRepo.CreateTable(ctx); err != nil {
		log.Fatalf("Failed to create stocks table: %v", err)
	}

	// Initialize service and handler
	stockService := service.NewStockService(cfg, stockRepo)
	stockHandler := handlers.NewStockHandler(stockService)

	// Configure CORS

	var router *gin.Engine = gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Data-Source", "X-Warning"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Stock Recommendation API is running!")
	})
	router.GET("/recommend", stockHandler.GetRecommendation)
	router.GET("/stocks", stockHandler.GetStocks)
	router.POST("/stocks/refresh", stockHandler.RefreshStocks)
	router.GET("/stock/:symbol", stockHandler.GetStockEOD)

	err = router.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
