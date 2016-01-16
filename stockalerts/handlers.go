package stockalerts

import (
	//	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"html/template"
	"net/http"
	"math"
)

var homeTemplate *template.Template

//ridiculous Go templates. Even for simple operations such as arithmatic, you need to define a function and map it
func round(num float64) int {
    return int(num + math.Copysign(0.5, num))
}
func toFixed(num float64, precision int) float64 {
    output := math.Pow(10, float64(precision))
    return float64(round(num * output)) / output
}

var funcMap = template.FuncMap{
	"netProfitLoss": netProfitLoss,
}

func netProfitLoss(boughtQuantity int, boughtPrice, lastTradedPrice float64) float64 {
	 f:=(float64(boughtQuantity) * lastTradedPrice) - (float64(boughtQuantity) * boughtPrice)
	return toFixed(f, 2)
}

func init() {
	r := mux.NewRouter()
	//	recoveryHandler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)) //for suppressing panics
	http.Handle("/", r)

	r.HandleFunc("/", rootHandler).Methods("GET")
	r.HandleFunc("/newUser", NewUser).Methods("POST")
	r.HandleFunc("/confirmUser", ConfirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", RegisterAlert).Methods("POST")
	r.HandleFunc("/loadStockPrices", LoadCurrentPrices).Methods("GET")
	r.HandleFunc("/removeAlert", RemoveAlert).Methods("Delete")
	r.HandleFunc("/getPortfolio", GetPortfolio).Methods("GET")
	r.HandleFunc("/updateAllPortfoliosAndAlert", UpdateAllPortfoliosAndAlert).Methods("GET")

	homeTemplate = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles("templates/home.html"))
}

type HomePageData struct {
	PortfolioAlert []PortfolioStock
	CachedStocks   map[string]Stock
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	alerts, _ := GetPortfolioStocksFor(ctx, "suman.jakkula@gmail.com")
	err := homeTemplate.Execute(w, HomePageData{PortfolioAlert: alerts, CachedStocks: cachedStocks})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
