package models

import "time"

type StockItem struct {
	Ticker     string    `json:"ticker"`
	TargetFrom string    `json:"target_from"`
	TargetTo   string    `json:"target_to"`
	Company    string    `json:"company"`
	Action     string    `json:"action"`
	Brokerage  string    `json:"brokerage"`
	RatingFrom string    `json:"rating_from"`
	RatingTo   string    `json:"rating_to"`
	Time       time.Time `json:"time"`
}

type StockResponse struct {
	Items    []StockItem `json:"items"`
	NextPage string      `json:"next_page"`
}

type Recommendation struct {
	Stock  StockItem `json:"stock"`
	Score  float64   `json:"score"`
	Reason string    `json:"reason"`
}
