package stockalerts

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"log"
)

//Key based add/update
func CreateOrUpdate(ctx context.Context, obj interface{}, kind string, stringId string, numericID int64) error {
	_, err := datastore.Put(ctx, datastore.NewKey(ctx, kind, stringId, numericID, nil), obj)
	if err != nil {
		log.Println("Failed to save object to datastore for kind:", kind, err)
		return err
	}
	return nil
}

//Key based retrieval
func GetEntity(ctx context.Context, stringId string, intId int64, kind string, entity interface{}) (err error) {
	if err = datastore.Get(ctx, datastore.NewKey(ctx, kind, stringId, intId, nil), entity); err != nil {
		log.Println("Did not find the entity with intId ", intId, "stringId ", stringId, " for kind = ", kind)
	}
	return
}

func LoadAllStockSymbols(ctx context.Context) []Stock {
	var stocks []Stock
	q := datastore.NewQuery("Stock")
	if _, err := q.GetAll(ctx, &stocks); err != nil {
		log.Println("Error in fetching all stocks ", err)
	}
	for _, s := range stocks {
		cachedStocks[s.Symbol] = s
	}
	return stocks
}
