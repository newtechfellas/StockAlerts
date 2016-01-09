package stockalerts

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	PublicApiUrl  = "http://query.yahooapis.com/v1/public/yql"
	DatatablesUrl = "store://datatables.org/alltableswithkeys"
)

func LoadCurrentPrices(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	if len(cachedStocks) == 0 {
		LoadAllStockSymbols(ctx)
	}
	// Even after loading all stocks, if stocks count is still 0, go home and have a beer
	// something is seriously wrong
	if len(cachedStocks) == 0 {
		log.Println("No stock symbols found in DB. Exiting")
		return
	}
	symbols := GetMapKeys(cachedStocks)
	stocks, err := Yql(ctx, symbols)
	if err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	JsonResponse(w, stocks, nil, http.StatusOK)
	return
}

type YqlJsonMeta struct {
	Count   int       `json:"count"`
	Created time.Time `json:"created"`
	Lang    string    `json:"lang"`
}

type YqlJsonQuoteResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Quote []Stock `json:"quote"`
		}
	}
}

type YqlJsonSingleQuoteResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Quote Stock `json:"quote"`
		}
	}
}

func Yql(ctx context.Context, symbols []string) (stocks []Stock, err error) {
	client := urlfetch.Client(ctx)

	quotedSymbols := MapStr(func(s string) string {
		return `"` + s + `"`
	}, symbols)

	query := fmt.Sprintf(`SELECT Symbol,Name,Open,LastTradePriceOnly,ChangeinPercent,DaysLow,DaysHigh,Change FROM %s WHERE symbol IN (%s)`,
		"yahoo.finance.quotes", strings.Join(quotedSymbols, ","))
	log.Println("Quotes query = ", query)

	v := url.Values{}
	v.Set("q", query)
	v.Set("format", "json")
	v.Set("env", DatatablesUrl)
	url := PublicApiUrl + "?" + v.Encode()
	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to fetch data from YQL for ", url, " error is ", err)
		return
	}
	defer resp.Body.Close()

	httpBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error in Reading from response body while fetching quote from YQL", err)
		return
	}
	log.Println("Response from YQL is ", string(httpBody[:]))
	if len(symbols) == 1 {
		var sresp YqlJsonSingleQuoteResponse
		if err = json.Unmarshal(httpBody, &sresp); err != nil {
			log.Println("Error in unmarshalling for single response ", err)
			return
		}
		stocks = append(stocks, sresp.Query.Results.Quote)
	} else {
		var resp YqlJsonQuoteResponse
		if err = json.Unmarshal(httpBody, &resp); err != nil {
			return
		}
		stocks = resp.Query.Results.Quote
	}
	return
}
