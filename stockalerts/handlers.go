package stockalerts

import (
//	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
"html/template"
	"google.golang.org/appengine"
)

var homeTemplate *template.Template

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

	homeTemplate = template.Must(template.ParseFiles("templates/home.html"))
}


func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	alerts, _ := GetPortfolioStocksFor(ctx,"suman.jakkula@gmail.com")
	err := homeTemplate.Execute(w,alerts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

