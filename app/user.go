package app

import (
	"log"
	"net/http"
)

type User struct {
	Username *string `json:"username"`
	FullName *string `json:"full_name"`
	GroupId  *int    `json:"group_id"`
	Role     *int    `json:"role"`
}

func getUserInfo(w http.ResponseWriter, r *http.Request, id int) {
	db, dbClose := openConnection()
	defer dbClose()

	err := db.Ping()
	if err != nil {
		log.Printf(err.Error())
		return
	}

	results, err := db.Query("SELECT `full_name`, `username`, `group_id`, `role` FROM `users` WHERE `id` = ?", id)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	var user User

	results.Next()

	err = results.Scan(&user.FullName, &user.Username, &user.GroupId, &user.Role)
	if err != nil {
		log.Printf(err.Error())
		return
	}

	json.NewEncoder(w).Encode(user)
}
