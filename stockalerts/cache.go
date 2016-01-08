package stockalerts

var cachedStockSymbols map[string]string = make(map[string]string, 500) //all known symbols to be cached. Map is used over slice to make faster lookups
