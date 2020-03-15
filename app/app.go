package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home")
}

func Run() error {

	readConfig()

	router := mux.NewRouter()

	router.HandleFunc("/", home)
	router.HandleFunc("/api/v1/auth/login", login).Methods("POST")
	router.HandleFunc("/api/v1/auth/register", register).Methods("POST")

	router.HandleFunc("/api/v1/group", isAuthenticated(getAllGroups)).Methods("GET")
	router.HandleFunc("/api/v1/group", isAuthenticated(createGroup)).Methods("POST")
	router.HandleFunc("/api/v1/group/{id}", isAuthenticated(getOneGroup)).Methods("GET")
	router.HandleFunc("/api/v1/group/{id}", isAuthenticated(updateGroup)).Methods("PATCH")
	router.HandleFunc("/api/v1/group/{id}", isAuthenticated(deleteGroup)).Methods("DELETE")

	log.Printf("Running at port 8080")
	return http.ListenAndServe(":8080", router)
}
