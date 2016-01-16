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

type PortfolioStock struct {
	Email          string `valid:"Required;"`
	Symbol         string `valid:"Required;"` //stock symbol such as AAPL. Works as unique identifier. Using this create the key
	PriceLow       float64
	PriceHigh      float64
	PriceBought    float64
	LastTradePrice float64
	AlertSentTime  time.Time `json:omitempty`
	CreatedTime    time.Time
	QuantityBought int
}

//stock symbols table
type Stock struct {
	Symbol          string
	Name            string
	Open            float64
	LastTradePrice  float64
	ChangeinPercent string
	DaysLow         float64
	DaysHigh        float64
	Change          string
	LastUpdated     time.Time
}

func (alert PortfolioStock) stringId() string {
	return alert.Email + ":" + alert.Symbol
}

func (alert PortfolioStock) isEligibleForAlert(s Stock) bool {
	//Alert is sent only once an hour
	//TODO: Make this user driven instead of hardcoded 1 hour limit
	return ((alert.PriceHigh != 0 && s.LastTradePrice > alert.PriceHigh) ||
		(alert.PriceLow != 0 && s.LastTradePrice < alert.PriceLow)) &&
		(alert.AlertSentTime.IsZero() || time.Now().Sub(alert.AlertSentTime).Hours() > 1)

}
func (alert PortfolioStock) kind() string {
	return "PortfolioStock"
}
func (s Stock) stringId() string {
	return s.Symbol
}
func (s Stock) kind() string {
	return "Stock"
}

func (f PortfolioStock) String() string { return Jsonify(f) }
func (u User) String() string           { return Jsonify(u) }
func (s Stock) String() string          { return Jsonify(s) }
