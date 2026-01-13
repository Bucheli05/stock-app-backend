package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bucheli05/stock-app-backend/internal/config"
	"github.com/Bucheli05/stock-app-backend/internal/models"
)

type StockService struct {
	cfg *config.Config
}

func NewStockService(cfg *config.Config) *StockService {
	return &StockService{cfg: cfg}
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
func (s *StockService) RecommendBestStock() (models.Recommendation, error) {
	stocks, err := s.FetchStocks()
	if err != nil {
		return models.Recommendation{}, err
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
