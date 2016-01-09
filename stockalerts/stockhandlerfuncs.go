package stockalerts

import (
	"encoding/json"
	"errors"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"time"
)

func removeAlert(w http.ResponseWriter, r *http.Request) {
	stockAlert := StockAlert{Symbol:r.URL.Query().Get("Symbol"), Email:r.URL.Query().Get("Email")}
	ctx := appengine.NewContext(r)
	log.Println("Removing stock alert ", stockAlert)
	if err := DeleteEntity(ctx,stockAlert.stringId(),0,stockAlert.kind()); err != nil {
		ErrorResponse(w,err,http.StatusInternalServerError)
		return
	}
	JsonResponse(w,nil,nil,http.StatusOK)
}
func registerAlert(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var stockAlert StockAlert
	if err := decoder.Decode(&stockAlert); err != nil {
		log.Println("Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Println("decoded stock alert:", stockAlert)
	if len(stockAlert.Email) == 0 || len(stockAlert.Symbol) == 0 ||
		(stockAlert.PriceLow == 0 && stockAlert.PriceHigh == 0) {
		log.Println("Invalid alert. Email, Symbol are mandatory. Either PriceLow or PriceHigh is mandatory")
		ErrorResponse(w, errors.New("Invalid json details in request body."), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	//Is user verified
	if _, err := GetValidUser(stockAlert.Email, ctx, w, r); err != nil {
		return
	}
	//Is stock symbol valid
	if _, ok := cachedStocks[stockAlert.Symbol]; !ok {
		if stocks , err := GetStocksUsingYql(ctx,[]string{ stockAlert.Symbol }); err != nil || len(stocks[0].Name) == 0 {
			log.Println("Invalid alert. Stock symbol ", stockAlert.Symbol, " does not exist")
			ErrorResponse(w, errors.New("Invalid alert. Stock symbol "+ stockAlert.Symbol+ " does not exist"), http.StatusBadRequest)
			return
		}
	}

	//Create stock alert entry
	stockAlert.CreatedTime = time.Now()
	if err := CreateOrUpdate(ctx, &stockAlert, stockAlert.kind(), stockAlert.stringId(), 0); err != nil {
		log.Println("Could not create stock alerts for ", stockAlert.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock alerts for "+stockAlert.Email), http.StatusInternalServerError)
		return
	}
	//If this is for a new stock, update stock table for this new symbol
	s := Stock{Symbol: stockAlert.Symbol}
	if err := CreateOrUpdate(ctx, &s, s.kind(), s.stringId(), 0); err != nil {
		log.Println("Could not create symbol ", stockAlert.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock symbol "+stockAlert.Email), http.StatusInternalServerError)
		return
	}
	cachedStocks[stockAlert.Symbol] = s //key and value are same. Weird!!!. But a map provides faster lookups
	//finally if all the above was successful return 202 Created status
	JsonResponse(w, nil, nil, http.StatusCreated)
	return
}