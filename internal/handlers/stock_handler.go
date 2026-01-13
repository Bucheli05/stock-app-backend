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
	recommendation, err := h.stockService.RecommendBestStock(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendation)
}

func (h *StockHandler) GetStocks(c *gin.Context) {
	// First try to get from database
	stocks, err := h.stockService.GetLatestStocksFromDB(c.Request.Context())
	if err != nil || len(stocks) == 0 {
		// If no stocks in DB, fetch from API and save
		stocks, fetchErr := h.stockService.FetchAndSaveStocks(c.Request.Context())
		if fetchErr != nil && len(stocks) == 0 {
			// Both API and DB failed with no data
			c.JSON(http.StatusInternalServerError, gin.H{"error": fetchErr.Error()})
			return
		}
		if fetchErr != nil && len(stocks) > 0 {
			// API failed but we have cached data - return it with a warning header
			c.Header("X-Data-Source", "cache")
			c.Header("X-Warning", "Using cached data - API temporarily unavailable")
		}
	}
	c.JSON(http.StatusOK, stocks)
}

// RefreshStocks forces a refresh from the API and saves to database
func (h *StockHandler) RefreshStocks(c *gin.Context) {
	stocks, err := h.stockService.FetchAndSaveStocks(c.Request.Context())
	if err != nil && len(stocks) == 0 {
		// Complete failure - no data available
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	if err != nil && len(stocks) > 0 {
		// API failed but returning cached data
		c.JSON(http.StatusOK, gin.H{
			"message":    "API unavailable - returning cached data",
			"count":      len(stocks),
			"warning":    err.Error(),
			"from_cache": true,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Stocks refreshed successfully",
		"count":   len(stocks),
	})
}
