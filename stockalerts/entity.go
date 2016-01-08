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
	PriceLow      float32
	PriceHigh     float32
	AlertSentTime time.Time
	CreatedTime   time.Time
}

func (alert StockAlert) getKey() string {
	return alert.Email + ":" + alert.Symbol
}

func (f StockAlert) String() string { return Jsonify(f) }
func (f User) String() string       { return Jsonify(f) }
