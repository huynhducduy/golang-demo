package app

import (
	"log"
	"net/http"
)

type Group struct {
	Id          *string `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func getAllGroups(w http.ResponseWriter, r *http.Request, id int) {

	db, dbClose := openConnection()
	defer dbClose()

	err := db.Ping()
	if err != nil {
		log.Printf(err.Error())
	}

	results, err := db.Query("SELECT `id`, `name`, `description` FROM `groups`")
	if err != nil {
		log.Printf(err.Error())
	}

	var groups []Group

	for results.Next() {
		var group Group

		err = results.Scan(&group.Id, &group.Name, &group.Description)
		if err != nil {
			log.Printf(err.Error())
		}

		groups = append(groups, group)

	}

	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request, id int) {
	// var newGroup Group
	// reqBody, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	// }

	// json.Unmarshal(reqBody, &newGroup)
	// groups = append(groups, newGroup)
	// w.WriteHeader(http.StatusCreated)

	// json.NewEncoder(w).Encode(newGroup)
}

func getOneGroup(w http.ResponseWriter, r *http.Request, id int) {
	// id := mux.Vars(r)["id"]

	// for _, group := range groups {
	// 	if group.Id == id {
	// 		json.NewEncoder(w).Encode(group)
	// 	}
	// }
}

func updateGroup(w http.ResponseWriter, r *http.Request, id int) {
	// id := mux.Vars(r)["id"]
	// var updatedGroup Group

	// reqBody, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	// }
	// json.Unmarshal(reqBody, &updatedGroup)

	// for i, group := range groups {
	// 	if group.Id == id {
	// 		group.Name = updatedGroup.Name
	// 		group.Description = updatedGroup.Description
	// 		groups = append(groups[:i], group)
	// 		json.NewEncoder(w).Encode(group)
	// 	}
	// }
}

func deleteGroup(w http.ResponseWriter, r *http.Request, id int) {
	// id := mux.Vars(r)["id"]

	// for i, group := range groups {
	// 	if group.Id == id {
	// 		groups = append(groups[:i], groups[i+1:]...)
	// 		fmt.Fprintf(w, "The event with ID %v has been deleted successfully", id)
	// 	}
	// }
}
