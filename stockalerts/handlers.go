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

	r.HandleFunc("/newUser", newUser).Methods("POST")
	r.HandleFunc("/confirmUser", confirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", registerAlert).Methods("POST")
	r.HandleFunc("/loadStockPrices", LoadCurrentPrices).Methods("POST")
	r.HandleFunc("/removeAlert", removeAlert).Methods("Delete")
}
