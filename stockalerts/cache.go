package stockalerts

var cachedStocks map[string]Stock = make(map[string]Stock, 500) //all known stocks to be cached. Map is used over slice to make faster lookups
