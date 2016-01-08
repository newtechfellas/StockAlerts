package stockalerts

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/newUser", newUser).Methods("POST")
	r.HandleFunc("/confirmUser", confirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", registerAlert).Methods("POST")
}
