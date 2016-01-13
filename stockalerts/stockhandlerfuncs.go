package stockalerts

import (
	"encoding/json"
	"errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"time"
)

func RemoveAlert(w http.ResponseWriter, r *http.Request) {
	if isTrustedReq(w, r) != nil {
		return
	}
	portfolioStock := PortfolioStock{Symbol: r.URL.Query().Get("symbol"), Email: r.URL.Query().Get("email")}
	ctx := appengine.NewContext(r)
	log.Debugf(ctx, "Removing stock alert ", portfolioStock)
	if err := DeleteEntity(ctx, portfolioStock.stringId(), 0, portfolioStock.kind()); err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	JsonResponse(w, nil, nil, http.StatusOK)
}

func GetPortfolio(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	ctx := appengine.NewContext(r)
	log.Debugf(ctx, "Fetching portfolio for ", email)
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
	ctx := appengine.NewContext(r)
	decoder := json.NewDecoder(r.Body)
	var portfolioStock PortfolioStock
	if err := decoder.Decode(&portfolioStock); err != nil {
		log.Debugf(ctx, "Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Debugf(ctx, "decoded stock alert:", portfolioStock)
	if len(portfolioStock.Email) == 0 || len(portfolioStock.Symbol) == 0 ||
		(portfolioStock.PriceLow == 0 && portfolioStock.PriceHigh == 0) {
		log.Debugf(ctx, "Invalid alert. Email, Symbol are mandatory. Either PriceLow or PriceHigh is mandatory")
		ErrorResponse(w, errors.New("Invalid json details in request body."), http.StatusBadRequest)
		return
	}
	//Is user verified
	if _, err := GetValidUser(portfolioStock.Email, ctx, w, r); err != nil {
		return
	}
	//Is stock symbol valid
	if _, ok := cachedStocks[portfolioStock.Symbol]; !ok {
		if stocks, err := GetStocksUsingYql(ctx, []string{portfolioStock.Symbol}); err != nil || len(stocks[0].Name) == 0 {
			log.Debugf(ctx, "Invalid alert. Stock symbol ", portfolioStock.Symbol, " does not exist")
			ErrorResponse(w, errors.New("Invalid alert. Stock symbol "+portfolioStock.Symbol+" does not exist"), http.StatusBadRequest)
			return
		} else {
			//New stock. Update cache and DB
			s := stocks[0]
			cachedStocks[s.Symbol] = s
			if err := CreateOrUpdate(ctx, &s, s.kind(), s.stringId(), 0); err != nil {
				log.Debugf(ctx, "Could not create symbol ", portfolioStock.Email, "Error is ", err)
				ErrorResponse(w, errors.New("Could not create stock symbol "+portfolioStock.Email), http.StatusInternalServerError)
				return
			}
		}
	}

	//Create stock alert entry
	portfolioStock.CreatedTime = time.Now()
	if err := CreateOrUpdate(ctx, &portfolioStock, portfolioStock.kind(), portfolioStock.stringId(), 0); err != nil {
		log.Debugf(ctx, "Could not create stock alerts for ", portfolioStock.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock alerts for "+portfolioStock.Email), http.StatusInternalServerError)
		return
	}
	//finally if all the above was successful return 202 Created status
	JsonResponse(w, nil, nil, http.StatusCreated)
	return
}
