package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// type Credentials struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

type allGroups []Group

var groups = allGroups{
	// {
	// 	Id:          "1",
	// 	Name:        "Introduction to Golang",
	// 	Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	// },
}

type Group struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home")
}

func login(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", config.DB_USER+":"+config.DB_PASS+"@tcp("+config.DB_HOST+":"+config.DB_PORT+")/"+config.DB_NAME)

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	results, err := db.Query("SELECT `name`, `description` FROM `groups`")
	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var group Group

		err = results.Scan(&group.Id, &group.Description)
		if err != nil {
			panic(err.Error())
		}

		log.Printf(group.Name)
		log.Printf(group.Description.String)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
}

func getAllGroups(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	var newGroup Group
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newGroup)
	groups = append(groups, newGroup)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newGroup)
}

func getOneGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	for _, group := range groups {
		if group.Id == id {
			json.NewEncoder(w).Encode(group)
		}
	}
}

func updateGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var updatedGroup Group

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}
	json.Unmarshal(reqBody, &updatedGroup)

	for i, group := range groups {
		if group.Id == id {
			group.Name = updatedGroup.Name
			group.Description = updatedGroup.Description
			groups = append(groups[:i], group)
			json.NewEncoder(w).Encode(group)
		}
	}
}

func deleteGroup(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	for i, group := range groups {
		if group.Id == id {
			groups = append(groups[:i], groups[i+1:]...)
			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", id)
		}
	}
}

type Config struct {
	DB_HOST string
	DB_PORT string
	DB_USER string
	DB_PASS string
	DB_NAME string
	SECRET  string
}

var config Config

func readConfig() {

	viper.SetConfigFile(".env")

	viper.AddConfigPath("..")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	log.Printf("DB_HOST %s", config.DB_HOST)
}

func Run() error {

	readConfig()

	router := mux.NewRouter()

	router.HandleFunc("/", home)
	router.HandleFunc("/api/v1/auth/login", login).Methods("POST")
	router.HandleFunc("/api/v1/auth/register", register).Methods("POST")

	router.HandleFunc("/api/v1/group", getAllGroups).Methods("GET")
	router.HandleFunc("/api/v1/group", createGroup).Methods("POST")
	router.HandleFunc("/api/v1/group/{id}", getOneGroup).Methods("GET")
	router.HandleFunc("/api/v1/group/{id}", updateGroup).Methods("PATCH")
	router.HandleFunc("/api/v1/group/{id}", deleteGroup).Methods("DELETE")

	fmt.Println("Running at port 8080")
	return http.ListenAndServe(":8080", router)
}
