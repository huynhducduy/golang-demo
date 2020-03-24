package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func Run() error {

	readConfig()

	openConnection()
	defer db.Close()

	router := mux.NewRouter()

	router.Use(loggingMiddleware)

	router.
		PathPrefix("/images/").
		Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("."+"/images/"))))

	router.HandleFunc("/api/v1/auth/login", login).Methods("POST")

	router.HandleFunc("/api/v1/group", isAuthenticated(getAllGroups)).Methods("GET")
	router.HandleFunc("/api/v1/group", isAuthenticated(createGroup)).Methods("POST")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(routerGetOneGroup)).Methods("GET")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(updateGroup)).Methods("PATCH")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}", isAuthenticated(deleteGroup)).Methods("DELETE")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}/member", isAuthenticated(getMembers)).Methods("GET")
	router.HandleFunc("/api/v1/group/addables", isAuthenticated(getAddableMembers)).Methods("GET")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}/member", isAuthenticated(addMember)).Methods("PUT")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}/member", isAuthenticated(setManager)).Methods("POST")
	router.HandleFunc("/api/v1/group/{id:[0-9]+}/member", isAuthenticated(removeMember)).Methods("DELETE")

	router.HandleFunc("/api/v1/task", isAuthenticated(getAllTasks)).Methods("GET")
	router.HandleFunc("/api/v1/task/assignable", isAuthenticated(getAssignableUsers)).Methods("GET")
	router.HandleFunc("/api/v1/task/reopenable", isAuthenticated(getReopenableTasks)).Methods("GET")
	router.HandleFunc("/api/v1/task", isAuthenticated(createTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(routerGetOneTask)).Methods("GET")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(updateTask)).Methods("PATCH")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}", isAuthenticated(deleteTask)).Methods("DELETE")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/check", isAuthenticated(checkTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/start", isAuthenticated(startTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/confirm", isAuthenticated(confirmTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/verify", isAuthenticated(verifyTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/close", isAuthenticated(closeTask)).Methods("POST")
	router.HandleFunc("/api/v1/task/{id:[0-9]+}/permission", isAuthenticated(getPermission)).Methods("GET")

	router.HandleFunc("/api/v1/user", isAuthenticated(getAllUsers)).Methods("GET")
	router.HandleFunc("/api/v1/user", isAuthenticated(createUser)).Methods("POST")
	router.HandleFunc("/api/v1/user/{id:[0-9]+}", isAuthenticated(routerGetOneUser)).Methods("GET")
	router.HandleFunc("/api/v1/user/{id:[0-9]+}", isAuthenticated(updateUser)).Methods("PATCH")
	router.HandleFunc("/api/v1/user/{id:[0-9]+}", isAuthenticated(deleteUser)).Methods("DELETE")

	router.HandleFunc("/api/v1/noti", isAuthenticated(getAllNotis)).Methods("GET")
	router.HandleFunc("/api/v1/noti", isAuthenticated(readAllNotis)).Methods("POST")
	router.HandleFunc("/api/v1/noti/{id:[0-9]+}", isAuthenticated(readNoti)).Methods("POST")

	router.HandleFunc("/api/v1/me", isAuthenticated(routerGetMe)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
	log.Printf("Running at port 8080")
	return http.ListenAndServe(":8080", c.Handler(handlers.RecoveryHandler()(handlers.CompressHandler(router))))
}
