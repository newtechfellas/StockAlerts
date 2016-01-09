package stockalerts

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/gorilla/handlers"
)

func init() {
	r := mux.NewRouter()
	recoveryHandler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)) //for suppressing panics
	http.Handle("/", recoveryHandler(r))

	r.HandleFunc("/newUser", newUser).Methods("POST")
	r.HandleFunc("/confirmUser", confirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", registerAlert).Methods("POST")
}
