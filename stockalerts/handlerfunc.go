package stockalerts

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	r := mux.NewRouter()
	http.Handle("/", r)
	r.HandleFunc("/newUser", newUser).Methods("POST")
	r.HandleFunc("/confirmUser", confirmUser).Methods("POST")
	r.HandleFunc("/registerAlert", registerAlert).Methods("POST")
}

func newUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	if err := decoder.Decode(&user); err != nil {
		log.Println("Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Println("decoded user:", user)
	if len(user.Email) == 0 {
		log.Println("User email is mandatory field")
		ErrorResponse(w, errors.New("User email is mandatory field"), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	if err := GetEntity(ctx, user.Email, 0, "User", &user); err == datastore.ErrNoSuchEntity {
		log.Println("Registering user", Jsonify(user))
		user.CreatedTime = time.Now()
		user.VerificationCode = rand.Int()
		if err = CreateOrUpdate(ctx, &user, "User", user.Email, 0); err != nil {
			log.Println("Error in creating user ", Jsonify(user), " error is ", err)
			ErrorResponse(w, errors.New("Error in creating user "), http.StatusInternalServerError)
		} else {
			//send email to confirm the verification code
			JsonResponse(w, nil, nil, http.StatusCreated)
		}
		return

	} else {
		//user already exists
		log.Println("Trying to register an existing user ", Jsonify(user))
		ErrorResponse(w, errors.New("User already exists "), http.StatusBadRequest)
		return
	}
}

func confirmUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	if err := decoder.Decode(&user); err != nil {
		log.Println("Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Println("decoded user:", user)
	if len(user.Email) == 0 {
		log.Println("User email is mandatory field")
		ErrorResponse(w, errors.New("User email is mandatory field"), http.StatusBadRequest)
		return
	}
	var dbUser User
	ctx := appengine.NewContext(r)
	if err := GetEntity(ctx, user.Email, 0, "User", &dbUser); err == datastore.ErrNoSuchEntity {
		log.Println("Trying to confirm a user that does not exist ", Jsonify(user))
		ErrorResponse(w, errors.New("Invalid request"), http.StatusBadRequest)
		return
	}

	if dbUser.VerificationCode == user.VerificationCode {
		log.Println("user confirmed ", user.Email)
		dbUser.IsVerified = true
		dbUser.VerifiedTime = time.Now()
		if err := CreateOrUpdate(ctx, &dbUser, "User", dbUser.Email, 0); err != nil {
			log.Println("Error in confirming user ", Jsonify(dbUser), " error is ", err)
			ErrorResponse(w, errors.New("Error in confirming user "), http.StatusInternalServerError)
		} else {
			//send email to confirm the verification code
			JsonResponse(w, nil, nil, http.StatusOK)
		}
		return
	} else {
		log.Println("Invalid verification code entered for user confirmation ", Jsonify(user))
		ErrorResponse(w, errors.New("Confirmation code is incorrect. Please check your email for the confirmation code and enter the correct value"), http.StatusOK)
	}
	return
}

func registerAlert(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var stockAlert StockAlert
	if err := decoder.Decode(&stockAlert); err != nil {
		log.Println("Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Println("decoded stock alert:", stockAlert)
	if len(stockAlert.Email) == 0 || len(stockAlert.Symbol) == 0 ||
		(stockAlert.PriceLow == 0 && stockAlert.PriceHigh == 0) {
		log.Println("Invalid alert. Email, Symbol are mandatory. Either PriceLow or PriceHigh is mandatory")
		ErrorResponse(w, errors.New("Invalid json details in request body."), http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	//Is user verified
	var user User
	if err := GetEntity(ctx, stockAlert.Email, 0, "User", &user); err != nil {
		log.Println("User not found for email ", stockAlert.Email)
		// do not return any error intentionally
		return
	}
	if !user.IsVerified {
		log.Println("User ", stockAlert.Email, " is not verified")
		ErrorResponse(w, errors.New("User is not verified. Check your email to confirm the registration"), http.StatusBadRequest)
		return
	}
	stockAlert.CreatedTime = time.Now()
	if err := CreateOrUpdate(ctx, &stockAlert, "StockAlert", stockAlert.getKey(), 0); err != nil {
		log.Println("Could not create stock alerts for ", stockAlert.Email, "Error is ", err)
		ErrorResponse(w, errors.New("Could not create stock alerts for "+stockAlert.Email), http.StatusInternalServerError)
		return
	}
	cachedStockSymbols[stockAlert.Symbol] = stockAlert.Symbol //key and value are same. Weird!!!. But a map is provides faster lookups
	//finally if all the above was successful return 202 Created status
	JsonResponse(w, nil, nil, http.StatusCreated)
	return
}
