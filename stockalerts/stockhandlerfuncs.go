package stockalerts

import (
	"encoding/json"
	"errors"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"time"
)

func RemoveAlert(w http.ResponseWriter, r *http.Request) {
	if isTrustedReq(w, r) != nil {
		return
	}
	portfolioStock := PortfolioStock{Symbol: r.URL.Query().Get("symbol"), Email: r.URL.Query().Get("email")}
	ctx := appengine.NewContext(r)
	log.Println("Removing stock alert ", portfolioStock)
	if err := DeleteEntity(ctx, portfolioStock.stringId(), 0, portfolioStock.kind()); err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	JsonResponse(w, nil, nil, http.StatusOK)
}

func GetPortfolio(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	ctx := appengine.NewContext(r)
	log.Println("Fetching portfolio for ", email)
	portfolioStocks, err := GetPortfolioStocksFor(ctx, email)
	if err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	JsonResponse(w, portfolioStocks, nil, http.StatusOK)
	return
}

func RegisterAlert(w http.ResponseWriter, r *http.Request) {
	if isTrustedReq(w, r) != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var portfolioStock PortfolioStock
	if err := decoder.Decode(&portfolioStock); err != nil {
		log.Println("Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Println("decoded stock alert:", portfolioStock)
	if len(portfolioStock.Email) == 0 || len(portfolioStock.Symbol) == 0 ||
		(portfolioStock.PriceLow == 0 && portfolioStock.PriceHigh == 0) {
		log.Println("Invalid alert. Email, Symbol are mandatory. Either PriceLow or PriceHigh is mandatory")
		ErrorResponse(w, errors.New("Invalid json details in request body."), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	//Is user verified
	if _, err := GetValidUser(portfolioStock.Email, ctx, w, r); err != nil {
		return
	}
	//Is stock symbol valid
	if _, ok := cachedStocks[portfolioStock.Symbol]; !ok {
		if stocks, err := GetStocksUsingYql(ctx, []string{portfolioStock.Symbol}); err != nil || len(stocks[0].Name) == 0 {
			log.Println("Invalid alert. Stock symbol ", portfolioStock.Symbol, " does not exist")
			ErrorResponse(w, errors.New("Invalid alert. Stock symbol "+portfolioStock.Symbol+" does not exist"), http.StatusBadRequest)
			return
		}
	}

	//Create stock alert entry
	portfolioStock.CreatedTime = time.Now()
	if err := CreateOrUpdate(ctx, &portfolioStock, portfolioStock.kind(), portfolioStock.stringId(), 0); err != nil {
		log.Println("Could not create stock alerts for ", portfolioStock.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock alerts for "+portfolioStock.Email), http.StatusInternalServerError)
		return
	}
	//If this is for a new stock, update stock table for this new symbol
	s := Stock{Symbol: portfolioStock.Symbol}
	if err := CreateOrUpdate(ctx, &s, s.kind(), s.stringId(), 0); err != nil {
		log.Println("Could not create symbol ", portfolioStock.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock symbol "+portfolioStock.Email), http.StatusInternalServerError)
		return
	}
	cachedStocks[portfolioStock.Symbol] = s //key and value are same. Weird!!!. But a map provides faster lookups
	//finally if all the above was successful return 202 Created status
	JsonResponse(w, nil, nil, http.StatusCreated)
	return
}
