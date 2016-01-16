package stockalerts

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"net/http"
	"strings"
	"time"
)

// This handler is invoked by scheduler every 1 min 30 seconds to compare against latest stock prices
// It sends email alerts if the configured portfolio stock is eligible for alert
// This handler does not load the latest stock prices from YQL. Instead it relies on the already loaded cache

// For any reason if scheduled run is not completed with-in 90 seconds, this variable prevents another run
var isUpdateInProgress bool = false

func isWeekDay() bool {
	today := time.Now().Weekday()
	switch today {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return true
	}
	return false
}

func UpdateAllPortfoliosAndAlert(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if len(cachedStocks) == 0 {
		log.Debugf(ctx, "UpdateAllPortfoliosAndAlert: cachedStocks are empty")
		return
	}
	if !isWeekDay() {
		log.Debugf(ctx, "UpdateAllPortfoliosAndAlert: is not a weekday.")
		return
	}
	defer func() { isUpdateInProgress = false }()
	isUpdateInProgress = true

	var portfolioStocks []PortfolioStock
	if err := GetAllEntities(ctx, "PortfolioStock", &portfolioStocks); err != nil {
		log.Debugf(ctx, "UpdateAllPortfoliosAndAlert: Could not fetch all portfolios ", err)
		return
	}
	if len(portfolioStocks) == 0 {
		log.Debugf(ctx, "UpdateAllPortfoliosAndAlert: fetched portfoilios len is 0")
		return
	}
	//	log.Debugf(ctx,"Before filteting eligible stocks ", Jsonify(portfolioStocks))
	eligibleStocks := FilterForAlertEligiblePortfolioStocks(ctx, portfolioStocks)
	//	log.Debugf(ctx,"After filteting eligible stocks ", Jsonify(portfolioStocks))
	SendAlerts(ctx, eligibleStocks)
}

func FilterForAlertEligiblePortfolioStocks(ctx context.Context, portfolioStocks []PortfolioStock) []PortfolioStock {
	var eligibleStocks []PortfolioStock
	for _, portfolioStock := range portfolioStocks {
		stock := cachedStocks[portfolioStock.Symbol]
		if portfolioStock.isEligibleForAlert(stock) {
			portfolioStock.LastTradePrice = stock.LastTradePrice
			eligibleStocks = append(eligibleStocks, portfolioStock)
		}
	}
	return eligibleStocks
}

func SendAlerts(ctx context.Context, stocksForAlert []PortfolioStock) {
	if len(stocksForAlert) == 0 {
		log.Debugf(ctx, "SendAlerts: Alert stocks are empty")
		return
	}
	//group by user emails to send a consolidated email
	groupedStockAlerts := make(map[string][]PortfolioStock)
	for _, alert := range stocksForAlert {
		userPortfolioStocks := groupedStockAlerts[alert.Email]
		userPortfolioStocks = append(userPortfolioStocks, alert)
		groupedStockAlerts[alert.Email] = userPortfolioStocks
	}
	log.Debugf(ctx, "Will send alerts for ", Jsonify(groupedStockAlerts))
	for email, alerts := range groupedStockAlerts {
		msg := &mail.Message{
			Sender:  "NewTechFellas Stock Alerts <newtechfellas@gmail.com>",
			To:      []string{email},
			Subject: "Newtechfellas stock alerts for your stocks",
			Body:    getStocksAlertMailBody(alerts),
		}
		if err := mail.Send(ctx, msg); err != nil {
			log.Debugf(ctx, "Couldn't send email: %v", err)
		}
	}

	for _, portfolioStock := range stocksForAlert {
		portfolioStock.AlertSentTime = time.Now()
		//Save stocksForAlert to update last alert sent time
		CreateOrUpdate(ctx, &portfolioStock, portfolioStock.kind(), portfolioStock.stringId(), 0)
	}
}
func getStocksAlertMailBody(portfolioStocks []PortfolioStock) string {
	var msgs []string
	for _, p := range portfolioStocks {
		msgs = append(msgs, fmt.Sprintf("Symbol: %v , Todays range %v - %v , Last traded at - %v ", p.Symbol, p.PriceLow, p.PriceHigh, p.LastTradePrice))
	}
	return strings.Join(msgs, "\n")
}
