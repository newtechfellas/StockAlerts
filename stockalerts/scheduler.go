package stockalerts

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/mail"
	"log"
	"net/http"
)

// This handler is invoked by scheduler every 1 min 30 seconds to compare against latest stock prices
// It sends email alerts if the configured portfolio stock is eligible for alert
// This handler does not load the latest stock prices from YQL. Instead it relies on the already loaded cache
// For any reason if scheduled run is not completed with-in 90 seconds, this variable prevents another run
var isUpdateInProgress bool = false

func UpdateAllPortfoliosAndAlert(w http.ResponseWriter, r *http.Request) {
	defer func() { isUpdateInProgress = false }()
	isUpdateInProgress = true
	if len(cachedStocks) == 0 {
		log.Println("Cached Stocks not available. Can not continue to update all portfolios")
		return
	}
	ctx := appengine.NewContext(r)
	var portfolioStocks []PortfolioStock
	if err := GetAllEntities(ctx, "PortfolioStock", &portfolioStocks); err != nil {
		log.Println("UpdateAllPortfoliosAndAlert: Could not fetch all portfolios ", err)
		return
	}
	if len(portfolioStocks) == 0 {
		log.Println("UpdateAllPortfoliosAndAlert: fetched portfoilios len is 0")
		return
	}
	eligibleStocks := FilterForAlertEligiblePortfolioStocks(portfolioStocks)
	SendAlerts(ctx, eligibleStocks)
}

func FilterForAlertEligiblePortfolioStocks(portfolioStocks []PortfolioStock) []PortfolioStock {
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
		log.Println("SendAlerts: Alert stocks are empty")
		return
	}
	//group by user emails to send a consolidated email
	groupedStockAlerts := make(map[string][]PortfolioStock)
	for _, alert := range stocksForAlert {
		userPortfolioStocks := groupedStockAlerts[alert.Email]
		userPortfolioStocks = append(userPortfolioStocks, alert)
		groupedStockAlerts[alert.Email] = userPortfolioStocks
	}
	log.Println("Will send alerts for ", Jsonify(groupedStockAlerts))
	for email, alerts := range groupedStockAlerts {
		msg := &mail.Message{
			Sender:  "NewTechFellas Stock Alerts <newtechfellas@gmail.com>",
			To:      []string{email},
			Subject: "Newtechfellas stock alerts for your stocks",
			Body:    Jsonify(alerts),
		}
		if err := mail.Send(ctx, msg); err != nil {
			log.Println(ctx, "Couldn't send email: %v", err)
		}
	}
}
