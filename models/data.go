package models

type StockData struct {
	Status     string  `json:"status"`
	From       string  `json:"from"`
	Symbol     string  `json:"symbol"`
	Date       string  `json:"date"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     float64 `json:"volume"`
	AfterHours float64 `json:"afterHours"`
	PreMarket  float64 `json:"preMarket"`
}
