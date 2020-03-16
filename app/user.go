package app

import (
	"errors"
	"log"
	"net/http"
)

type User struct {
	Id       *int    `json:"id,omitempty"`
	Username *string `json:"username"`
	FullName *string `json:"full_name"`
	GroupId  *int    `json:"group_id"`
	IsAdmin  *bool   `json:"is_admin"`
}

func getMe(id int) (*User, error) {
	db, dbClose, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `full_name`, `username`, `group_id`, `is_admin` FROM `users` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	var user User

	if !results.Next() {

		return nil, errors.New("Invalid user id")
	} else {
		err = results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId, &user.IsAdmin)
		if err != nil {
			return nil, err
		}

		return &user, nil
	}
}

func routerGetMe(w http.ResponseWriter, r *http.Request, user User) {
	json.NewEncoder(w).Encode(user)
}

func isManager(id int) (bool, error) {

	db, dbClose, err := openConnection()
	if err != nil {
		return false, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `manager_id` FROM `groups` WHERE `id` = (SELECT `group_id` FROM `users` where `id` = ?)", id)
	if err != nil {
		log.Printf(err.Error())
		return false, nil
	}

	if results.Next() {

		err = results.Scan(&id)
		if err != nil {
			log.Printf(err.Error())
			return false, nil
		}

		var manager_id int

		results.Scan(&manager_id)

		if manager_id == id {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
