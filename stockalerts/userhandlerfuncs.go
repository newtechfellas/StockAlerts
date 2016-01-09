package stockalerts

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/mail"
	"google.golang.org/appengine/log"
	"math/rand"
	"net/http"
	"time"
)

func NewUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	ctx := appengine.NewContext(r)
	var user User
	if err := decoder.Decode(&user); err != nil {
		log.Debugf(ctx, "Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Debugf(ctx, "decoded user:", user)
	if len(user.Email) == 0 {
		log.Debugf(ctx, "User email is mandatory field")
		ErrorResponse(w, errors.New("User email is mandatory field"), http.StatusBadRequest)
		return
	}
	if err := GetEntity(ctx, user.Email, 0, "User", &user); err == datastore.ErrNoSuchEntity {
		log.Debugf(ctx, "Registering user", Jsonify(user))
		user.CreatedTime = time.Now()
		user.VerificationCode = rand.Int()
		if err = CreateOrUpdate(ctx, &user, "User", user.Email, 0); err != nil {
			log.Debugf(ctx, "Error in creating user ", Jsonify(user), " error is ", err)
			ErrorResponse(w, errors.New("Error in creating user "), http.StatusInternalServerError)
		} else {
			//send email to confirm the verification code
			sendVerificationCodeEmail(ctx, user)
			JsonResponse(w, nil, nil, http.StatusCreated)
		}
		return

	} else {
		//user already exists
		log.Debugf(ctx, "Trying to register an existing user ", Jsonify(user))
		ErrorResponse(w, errors.New("User already exists "), http.StatusBadRequest)
		return
	}
}

func ConfirmUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	ctx := appengine.NewContext(r)
	if err := decoder.Decode(&user); err != nil {
		log.Debugf(ctx, "Error in decoding request body. Error is ", err)
		ErrorResponse(w, errors.New("Invalid json details in request body"), http.StatusBadRequest)
		return
	}
	log.Debugf(ctx, "decoded user:", user)
	if len(user.Email) == 0 {
		log.Debugf(ctx, "User email is mandatory field")
		ErrorResponse(w, errors.New("User email is mandatory field"), http.StatusBadRequest)
		return
	}
	var dbUser User
	if err := GetEntity(ctx, user.Email, 0, "User", &dbUser); err == datastore.ErrNoSuchEntity {
		log.Debugf(ctx, "Trying to confirm a user that does not exist ", Jsonify(user))
		ErrorResponse(w, errors.New("Invalid request"), http.StatusBadRequest)
		return
	}

	if dbUser.VerificationCode == user.VerificationCode {
		log.Debugf(ctx, "user confirmed ", user.Email)
		dbUser.IsVerified = true
		dbUser.VerifiedTime = time.Now()
		if err := CreateOrUpdate(ctx, &dbUser, "User", dbUser.Email, 0); err != nil {
			log.Debugf(ctx, "Error in confirming user ", Jsonify(dbUser), " error is ", err)
			ErrorResponse(w, errors.New("Error in confirming user "), http.StatusInternalServerError)
		} else {
			//send email to confirm the verification code
			JsonResponse(w, nil, nil, http.StatusOK)
		}
		return
	} else {
		log.Debugf(ctx, "Invalid verification code entered for user confirmation ", Jsonify(user))
		ErrorResponse(w, errors.New("Confirmation code is incorrect. Please check your email for the confirmation code and enter the correct value"), http.StatusOK)
	}
	return
}

func GetValidUser(email string, ctx context.Context, w http.ResponseWriter, r *http.Request) (user User, err error) {
	//Is user verified
	if err = GetEntity(ctx, email, 0, "User", &user); err != nil {
		log.Debugf(ctx, "User not found for email ", email)
		return
	}
	if !user.IsVerified {
		log.Debugf(ctx, "User ", email, " is not verified")
		ErrorResponse(w, errors.New("User is not verified. Check your email to confirm the registration"), http.StatusBadRequest)
		return
	}
	return
}

func sendVerificationCodeEmail(ctx context.Context, user User) {
	msg := &mail.Message{
		Sender:  "NewTechFellas Stock Alerts Admin <newtechfellas@gmail.com>",
		To:      []string{user.Email},
		Subject: "Newtechfellas stock alerts verify user",
		Body:    fmt.Sprintf("Your confirmation code is %v", user.VerificationCode),
	}
	if err := mail.Send(ctx, msg); err != nil {
		log.Debugf(ctx, "Couldn't send email: %v", err)
	}
}
