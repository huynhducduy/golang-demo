package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Payload struct {
	Id         int
	Expires_at int64
	jwt.StandardClaims
}

type Token struct {
	Token      string `json:"token"`
	Expires_at int64  `json:"expires_at"`
}

func getToken(r *http.Request) (string, error) {
	if r.Header["Authorization"] != nil && len(strings.Split(r.Header["Authorization"][0], " ")) == 2 {
		return strings.Split(r.Header["Authorization"][0], " ")[1], nil
	} else {
		return "", errors.New("No bearer token")
	}
}

func isAuthenticated(endpoint func(http.ResponseWriter, *http.Request, User)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		user_token, err := getToken(r)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		token, err := jwt.ParseWithClaims(user_token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SECRET), nil
		})

		if err != nil {
			responseInternalError(w, err)
			return
		}

		if token.Valid {

			db, dbClose, err := openConnection()
			if err != nil {
				responseInternalError(w, err)
				return
			}
			defer dbClose()

			var user User
			user.Id = &token.Claims.(*Payload).Id

			results, err := db.Query("SELECT `full_name`, `username`, `group_id`, `role` FROM `users` WHERE `id` = ?", user.Id)
			if err != nil {
				responseInternalError(w, err)
				return
			}

			results.Next()

			err = results.Scan(&user.FullName, &user.Username, &user.GroupId, &user.IsAdmin)
			if err != nil {
				responseInternalError(w, err)
				return
			}

			endpoint(w, r, user)
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
		responseInternalError(w, err)
		return
	}

	json.Unmarshal(reqBody, &credential)

	if credential.Username != "" && credential.Password != "" {
		db, dbClose, err := openConnection()
		if err != nil {
			responseInternalError(w, err)
			return
		}
		defer dbClose()

		results, err := db.Query("SELECT `id` FROM `users` where `username` = ? AND `password` = ?", credential.Username, credential.Password)
		if err != nil {
			responseInternalError(w, err)
			return
		}

		if results.Next() {
			var id int

			err = results.Scan(&id)
			if err != nil {
				responseCustomError(w, err, http.StatusNotFound, "Username and passowrd is incorrect!")
				return
			}

			token := generateToken(id)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(token)
		}

		responseCustomError(w, err, http.StatusNotFound, "Username and passowrd is incorrect!")
		return
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
