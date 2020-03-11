package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// var jwtKey = []byte("ahihi do ngok")

// type user struct {
// 	id       string `json:"id"`
// 	name     string `json:"name"`
// 	password string `json:"password"`
// 	group_id string `json:"group_id"`
// 	role     string `json:"role"`
// }

// type allUsers []user

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome homea!")
}

func login(w http.ResponseWriter, r *http.Request) {
	_, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	// var Credentials credentials

	// json.Unmarshal(reqBody, &credentials)
	// w.WriteHeader(http.StatusCreated)

	// json.NewEncoder(w).Encode(credentials)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", home)
	router.HandleFunc("api/v1/auth/login", login)

	fmt.Println("Running at port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
