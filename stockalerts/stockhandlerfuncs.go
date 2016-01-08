package stockalerts

import (
	"encoding/json"
	"errors"
	"google.golang.org/appengine"
	"log"
	"net/http"
	"time"
)

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
	stockAlert.CreatedTime = time.Now()
	if err := CreateOrUpdate(ctx, &stockAlert, "StockAlert", stockAlert.getKey(), 0); err != nil {
		log.Println("Could not create stock alerts for ", stockAlert.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock alerts for "+stockAlert.Email), http.StatusInternalServerError)
		return
	}
	cachedStockSymbols[stockAlert.Symbol] = stockAlert.Symbol //key and value are same. Weird!!!. But a map is provides faster lookups
	//finally if all the above was successful return 202 Created status
	JsonResponse(w, nil, nil, http.StatusCreated)
	return
}
