package stockalerts

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

//Key based add/update
func CreateOrUpdate(ctx context.Context, obj interface{}, kind string, stringId string, numericID int64) error {
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, kind, stringId, numericID, nil), obj)
	if err != nil {
		log.Debugf(ctx, "Failed to save object to datastore for kind:", kind, err)
		return err
	}
	return nil
}

//Key based retrieval
func GetEntity(ctx context.Context, stringId string, intId int64, kind string, entity interface{}) (err error) {
	if err = datastore.Get(ctx, datastore.NewKey(ctx, kind, stringId, intId, nil), entity); err != nil {
		log.Debugf(ctx, "Did not find the entity with intId ", intId, "stringId ", stringId, " for kind = ", kind)
	}
	return
}

func GetAllEntities(ctx context.Context, kind string, entities interface{}) (err error) {
	q := datastore.NewQuery(kind)
	_, err = q.GetAll(ctx, entities)
	return
}

//Key based retrieval
func DeleteEntity(ctx context.Context, stringId string, intId int64, kind string) (err error) {
	if err = datastore.Delete(ctx, datastore.NewKey(ctx, kind, stringId, intId, nil)); err != nil {
		log.Debugf(ctx, "Did not find the entity with intId ", intId, "stringId ", stringId, " for kind = ", kind)
	}
	return
}

func GetPortfolioStocksFor(ctx context.Context, email string) (alerts []PortfolioStock, err error) {
	q := datastore.NewQuery("PortfolioStock").Filter("Email =", email)
	if _, err = q.GetAll(ctx, &alerts); err != nil {
		log.Debugf(ctx, "Could not fetch stock alerts for email ", email)
		return
	}
	//Update the portfolio alerts
	if len(cachedStocks) == 0 {
		log.Debugf(ctx, "Stocks current prices are not available yet. Perhaps the scheduler has not begun or failed")
		return
	}
	log.Debugf(ctx, "updating last traded price using cachedStocks")
	for symbol, stock := range cachedStocks {
		for index, portfolioStock := range alerts {
			if portfolioStock.Symbol == symbol {
				portfolioStock.LastTradePrice = stock.LastTradePrice
			}
			alerts[index] = portfolioStock
		}
	}
	log.Debugf(ctx, "Returning ", len(alerts), " number of alerts for email ", email)
	return
}

func LoadAllStockSymbols(ctx context.Context) []Stock {
	var stocks []Stock
	q := datastore.NewQuery("Stock")
	if _, err := q.GetAll(ctx, &stocks); err != nil {
		log.Debugf(ctx, "Error in fetching all stocks ", err)
	}
	for _, s := range stocks {
		cachedStocks[s.Symbol] = s
	}
	return stocks
}
