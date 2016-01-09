package stockalerts

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	r := mux.NewRouter()
	recoveryHandler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)) //for suppressing panics
	http.Handle("/", recoveryHandler(r))

	r.HandleFunc("/newUser", NewUser).Methods("POST")
	r.HandleFunc("/confirmUser", ConfirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", RegisterAlert).Methods("POST")
	r.HandleFunc("/loadStockPrices", LoadCurrentPrices).Methods("GET")
	r.HandleFunc("/removeAlert", RemoveAlert).Methods("Delete")
	r.HandleFunc("/getPortfolio", GetPortfolio).Methods("GET")
	r.HandleFunc("/updateAllPortfoliosAndAlert", UpdateAllPortfoliosAndAlert).Methods("GET")
}
