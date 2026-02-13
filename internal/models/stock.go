package models

import "time"

type StockItem struct {
	ID         int       `json:"id" db:"id"`
	Ticker     string    `json:"ticker" db:"ticker"`
	TargetFrom string    `json:"target_from" db:"target_from"`
	TargetTo   string    `json:"target_to" db:"target_to"`
	Company    string    `json:"company" db:"company"`
	Action     string    `json:"action" db:"action"`
	Brokerage  string    `json:"brokerage" db:"brokerage"`
	RatingFrom string    `json:"rating_from" db:"rating_from"`
	RatingTo   string    `json:"rating_to" db:"rating_to"`
	Time       time.Time `json:"time" db:"time"`
	FetchedAt  time.Time `json:"fetched_at" db:"fetched_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
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

// MarketStack EOD Data Models
type EODData struct {
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      float64 `json:"volume"`
	AdjHigh     float64 `json:"adj_high"`
	AdjLow      float64 `json:"adj_low"`
	AdjClose    float64 `json:"adj_close"`
	AdjOpen     float64 `json:"adj_open"`
	AdjVolume   float64 `json:"adj_volume"`
	SplitFactor float64 `json:"split_factor"`
	Dividend    float64 `json:"dividend"`
	Symbol      string  `json:"symbol"`
	Exchange    string  `json:"exchange"`
	Date        string  `json:"date"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

type MarketStackResponse struct {
	Pagination Pagination `json:"pagination"`
	Data       []EODData  `json:"data"`
}
