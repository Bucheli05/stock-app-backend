package handlers

import (
	"net/http"

	"github.com/Bucheli05/stock-app-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockService *service.StockService
}

func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

func (h *StockHandler) GetRecommendation(c *gin.Context) {
	recommendation, err := h.stockService.RecommendBestStock()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}
