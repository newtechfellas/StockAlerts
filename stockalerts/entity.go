package stockalerts

import (
	"time"
)

type User struct {
	Name             string //optional
	Email            string `valid:"Required;"` //unique identifier
	VerificationCode int
	IsVerified       bool
	PhoneNumber      string //optional
	CreatedTime      time.Time
	VerifiedTime     time.Time
}

type StockAlert struct {
	Email         string `valid:"Required;"`
	Symbol        string `valid:"Required;"` //stock symbol such as AAPL. Works as unique identifier. Using this create the key
	PriceLow      float64
	PriceHigh     float64
	AlertSentTime time.Time
	CreatedTime   time.Time
}

//stock symbols table
type Stock struct {
	Symbol             string
	Name               string
	Open               float64
	LastTradePrice float64
	ChangeinPercent    string
	DaysLow            float64
	DaysHigh           float64
	Change             string
	LastUpdated        time.Time
}

func (alert StockAlert) stringId() string {
	return alert.Email + ":" + alert.Symbol
}
func (alert StockAlert) kind() string {
	return "StockAlert"
}
func (s Stock) stringId() string {
	return s.Symbol
}
func (s Stock) kind() string {
	return "Stock"
}

func (f StockAlert) String() string { return Jsonify(f) }
func (u User) String() string       { return Jsonify(u) }
func (s Stock) String() string      { return Jsonify(s) }
