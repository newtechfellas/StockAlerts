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
	"strconv"
)

const (
	PublicApiUrl  = "http://query.yahooapis.com/v1/public/yql"
	DatatablesUrl = "store://datatables.org/alltableswithkeys"
)

type YqlJsonMeta struct {
	Count   int       `json:"count"`
	Created time.Time `json:"created"`
	Lang    string    `json:"lang"`
}

type Quote struct {
	Symbol             string
	Name               string
	Open               string
	LastTradePriceOnly string
	ChangeinPercent    string
	DaysLow            string
	DaysHigh           string
	Change             string
}

type YqlJsonQuoteResponse struct {
	Query struct {
			  YqlJsonMeta
			  Results struct {
						  Quote []Quote `json:"quote"`
					  }
		  }
}

type YqlJsonSingleQuoteResponse struct {
	Query struct {
			  YqlJsonMeta
			  Results struct {
						  Quote Quote `json:"quote"`
					  }
		  }
}

func (q Quote )toStock() Stock {
	var s Stock
	s.Name = q.Name
	s.Symbol = q.Symbol
	s.ChangeinPercent = q.ChangeinPercent
	s.Open, _ = strconv.ParseFloat(q.Open,64)
	s.LastTradePrice, _ = strconv.ParseFloat(q.LastTradePriceOnly,64)
	s.DaysHigh, _ = strconv.ParseFloat(q.DaysHigh,64)
	s.DaysLow, _ = strconv.ParseFloat(q.DaysLow,64)
	s.Change = q.Change
	return s
}

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
	_, err := GetStocksUsingYql(ctx, symbols)
	if err != nil {
		ErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	log.Println("cachedStocks = ",Jsonify(cachedStocks))
	JsonResponse(w, cachedStocks, nil, http.StatusOK)
	return
}

func GetStocksUsingYql(ctx context.Context, symbols []string) (stocks []Stock, err error) {
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
		stocks = append(stocks, sresp.Query.Results.Quote.toStock())
	} else {
		var resp YqlJsonQuoteResponse
		if err = json.Unmarshal(httpBody, &resp); err != nil {
			return
		}
		for _,q := range resp.Query.Results.Quote {
			stocks = append(stocks,q.toStock())
		}
	}
	for _,s := range stocks {
		s.LastUpdated = time.Now()
		cachedStocks[s.Symbol] = s //update the cache
	}
	return
}