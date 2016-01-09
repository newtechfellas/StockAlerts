package stockalerts

import (
	"google.golang.org/appengine"
	"log"
	"net/http"
)

func LoadCurrentPrices(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if len(cachedStocks) == 0 {
		_ = LoadAllStockSymbols(ctx)
	}
	if len(cachedStocks) == 0 {
		log.Println("No stock symbols found in DB. Exiting")
		return
	}
}
