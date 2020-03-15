package app

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Payload struct {
	Id         int
	Expires_at int64
	jwt.StandardClaims
}

type Token struct {
	Token      string `json:"token"`
	Expires_at int64  `json:"expires_at"`
}

func isAuthenticated(endpoint func(http.ResponseWriter, *http.Request, int)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Authorization"] != nil && len(strings.Split(r.Header["Authorization"][0], " ")) == 2 {
			user_token := strings.Split(r.Header["Authorization"][0], " ")[1]

			token, err := jwt.ParseWithClaims(user_token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.SECRET), nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				log.Printf(err.Error())
				return
			}

			if token.Valid {
				endpoint(w, r, token.Claims.(*Payload).Id)
			}

		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
}

func generateToken(id int) Token {
	exp_at := time.Now().Unix() + 604800 // 1 week

	payload := Payload{Id: id, Expires_at: exp_at}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), payload)
	tokenString, _ := token.SignedString([]byte(config.SECRET))

	return Token{
		Token:      tokenString,
		Expires_at: exp_at,
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var credential Credential
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err.Error())
	}

	json.Unmarshal(reqBody, &credential)

	if credential.Username != "" && credential.Password != "" {
		db, dbClose := openConnection()
		defer dbClose()

		results, err := db.Query("SELECT `id` FROM `users` where `username` = ? AND `password` = ?", credential.Username, credential.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf(err.Error())
			return
		}

		results.Next()
		var id int

		err = results.Scan(&id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			log.Printf(err.Error())
			return
		}

		token := generateToken(id)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(token)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func register(w http.ResponseWriter, r *http.Request) {
}
