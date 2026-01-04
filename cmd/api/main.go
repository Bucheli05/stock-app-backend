package main

import (
	"fmt"
	"log"

	"github.com/Bucheli05/stock-app-backend/internal/config"
	"github.com/Bucheli05/stock-app-backend/internal/handlers"
	"github.com/Bucheli05/stock-app-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	stockService := service.NewStockService(cfg)
	stockHandler := handlers.NewStockHandler(stockService)

	var router *gin.Engine = gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Stock Recommendation API is running!")
	})

	router.GET("/recommend", stockHandler.GetRecommendation)

	err = router.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
