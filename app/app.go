package app

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Run() error {

	readConfig()

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/auth/login", login).Methods("POST")

	router.HandleFunc("/api/v1/group", isAuthenticated(getAllGroups)).Methods("GET")
	router.HandleFunc("/api/v1/group", isAuthenticated(createGroup)).Methods("POST")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(getOneGroup)).Methods("GET")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(updateGroup)).Methods("PATCH")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(deleteGroup)).Methods("DELETE")

	router.HandleFunc("/api/v1/task", isAuthenticated(getAllTasks)).Methods("GET")
	router.HandleFunc("/api/v1/task/assignable", isAuthenticated(listAssignableUsers)).Methods("GET")
	router.HandleFunc("/api/v1/task", isAuthenticated(createTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(routerGetOneTask)).Methods("GET")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(updateTask)).Methods("PATCH")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(deleteTask)).Methods("DELETE")

	router.HandleFunc("/api/v1/me", isAuthenticated(routerGetMe)).Methods("GET")

	log.Printf("Running at port 8080")
	return http.ListenAndServe(":8080", router)
}
