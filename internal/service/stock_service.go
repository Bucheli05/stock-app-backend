package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Bucheli05/stock-app-backend/internal/config"
	"github.com/Bucheli05/stock-app-backend/internal/models"
	"github.com/Bucheli05/stock-app-backend/internal/repository"
)

type StockService struct {
	cfg  *config.Config
	repo *repository.StockRepository
}

func NewStockService(cfg *config.Config, repo *repository.StockRepository) *StockService {
	return &StockService{
		cfg:  cfg,
		repo: repo,
	}
}

// FetchStocks fetches the list of stocks from the API.
func (s *StockService) FetchStocks() ([]models.StockItem, error) {
	var allStocks []models.StockItem
	nextPage := ""

	for {
		stocks, next, err := s.fetchStocksPage(nextPage)
		if err != nil {
			return nil, err
		}

		allStocks = append(allStocks, stocks...)

		// If there's no next page, we're done
		if next == "" {
			break
		}
		nextPage = next
	}

	return allStocks, nil
}

// FetchAndSaveStocks fetches stocks from API and saves them to the database
// If API fails, returns cached data from database as fallback
func (s *StockService) FetchAndSaveStocks(ctx context.Context) ([]models.StockItem, error) {
	// Fetch stocks from API
	stocks, err := s.FetchStocks()
	if err != nil {
		// API failed, try to return cached data from database
		cachedStocks, dbErr := s.repo.GetLatestStocks(ctx)
		if dbErr == nil && len(cachedStocks) > 0 {
			// Return cached data with a note that it's from cache
			return cachedStocks, fmt.Errorf("API unavailable, returning cached data: %w", err)
		}
		// No cached data available
		return nil, fmt.Errorf("failed to fetch stocks and no cached data available: %w", err)
	}

	// Save to database with current timestamp
	fetchedAt := time.Now()
	if err := s.repo.SaveStocks(ctx, stocks, fetchedAt); err != nil {
		// If save fails, still return the fetched data
		return stocks, fmt.Errorf("stocks fetched but failed to save to database: %w", err)
	}

	return stocks, nil
}

// GetLatestStocksFromDB retrieves the most recent stocks from the database
func (s *StockService) GetLatestStocksFromDB(ctx context.Context) ([]models.StockItem, error) {
	return s.repo.GetLatestStocks(ctx)
}

// FetchMarketStackEOD fetches end-of-day data from MarketStack API for a given symbol
func (s *StockService) FetchMarketStackEOD(symbol string) (*models.EODData, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build the MarketStack API URL
	url := fmt.Sprintf("http://api.marketstack.com/v1/eod?access_key=5ff4770718f41c8dfce60791d0e75749&symbols=%s&exchange=NASDAQ&limit=1", symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from MarketStack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MarketStack API request failed with status: %d", resp.StatusCode)
	}

	var marketStackResp models.MarketStackResponse
	if err := json.NewDecoder(resp.Body).Decode(&marketStackResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(marketStackResp.Data) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	return &marketStackResp.Data[0], nil
}

// fetchStocksPage fetches a single page of stocks from the API.
func (s *StockService) fetchStocksPage(nextPage string) ([]models.StockItem, string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", s.cfg.ApiUrl, nil)
	if err != nil {
		return nil, "", err
	}

	// Add next_page parameter if it exists
	if nextPage != "" {
		q := req.URL.Query()
		q.Add("next_page", nextPage)
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Add("Authorization", "Bearer "+s.cfg.AuthToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var response models.StockResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, "", err
	}

	return response.Items, response.NextPage, nil
}

// parsePrice converts a string like "$428.00" to a float64.
func parsePrice(priceStr string) (float64, error) {
	priceStr = strings.TrimPrefix(priceStr, "$")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, err // or handle appropriately
	}
	return price, nil
}

// RecommendBestStock analyzes the stocks and returns the best one.
func (s *StockService) RecommendBestStock(ctx context.Context) (models.Recommendation, error) {
	// Try to get stocks from database first
	stocks, err := s.GetLatestStocksFromDB(ctx)
	if err != nil || len(stocks) == 0 {
		// If no stocks in DB, try to fetch from API and save
		stocks, err = s.FetchAndSaveStocks(ctx)
		if err != nil && len(stocks) == 0 {
			// Both API and DB failed
			return models.Recommendation{}, err
		}
		// If we got here, we have cached data even though API failed
	}

	var bestStock models.StockItem
	var maxScore float64 = -1e9
	var bestReason string

	for _, stock := range stocks {
		// Only consider stocks where the target was raised
		if !strings.Contains(strings.ToLower(stock.Action), "raised") {
			continue
		}

		from, _ := parsePrice(stock.TargetFrom)
		to, _ := parsePrice(stock.TargetTo)

		if from == 0 {
			continue
		}

		// Score is the percentage increase in target price
		score := ((to - from) / from) * 100

		// Bonus for "Buy" or "Outperform" ratings
		if strings.EqualFold(stock.RatingTo, "Buy") || strings.EqualFold(stock.RatingTo, "Outperform") {
			score += 5.0
		}

		if score > maxScore {
			maxScore = score
			bestStock = stock
			bestReason = fmt.Sprintf("Highest target price increase (%.2f%%) with a positive action and rating.", score)
		}
	}

	if bestStock.Ticker == "" {
		return models.Recommendation{}, fmt.Errorf("no suitable stock recommendations found")
	}

	return models.Recommendation{
		Stock:  bestStock,
		Score:  maxScore,
		Reason: bestReason,
	}, nil
}
