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

func getOneUser(id int) (*User, error) {
	db, dbClose, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `id`, `full_name`, `username`, `group_id`, `is_admin` FROM `users` WHERE `id` = ?", id)
	if err != nil {
		return nil, err
	}

	if results.Next() {
		var user User

		err = results.Scan(&user.Id, &user.FullName, &user.Username, &user.GroupId, &user.IsAdmin)
		if err != nil {
			return nil, err
		}

		return &user, nil
	} else {
		return nil, errors.New("Invalid user id")
	}
}

func routerGetMe(w http.ResponseWriter, r *http.Request, user User) {
	json.NewEncoder(w).Encode(user)
}

func isManagerOf(groupId int, userId int) (bool, error) {

	db, dbClose, err := openConnection()
	if err != nil {
		return false, err
	}
	defer dbClose()

	results, err := db.Query("SELECT `manager_id` FROM `groups` WHERE `id` = ", groupId)
	if err != nil {
		log.Printf(err.Error())
		return false, nil
	}

	if results.Next() {

		var manager_id int

		results.Scan(&manager_id)

		if manager_id == userId {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, nil
}
