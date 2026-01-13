package models

import "time"

type TickerResponse struct {
	Data []TickerItem `json:"data"`
}

type TickerItem struct {
	Open          float64   `json:"open"`
	Close         float64   `json:"close"`
	PriceCurrency string    `json:"price_currency"`
	Symbol        string    `json:"symbol"`
	Date          time.Time `json:"date"`
}
