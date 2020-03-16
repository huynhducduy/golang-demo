package app

import (
	"log"
	"net/http"
)

type User struct {
	Id       *int    `json:"id,omitempty"`
	Username *string `json:"username"`
	FullName *string `json:"full_name"`
	GroupId  *int    `json:"group_id"`
	Role     *int    `json:"role"`
}

func getMe(w http.ResponseWriter, r *http.Request, user User) {
	db, dbClose := openConnection()
	defer dbClose()

	err := db.Ping()
	if err != nil {
		log.Printf(err.Error())
		return
	}

	json.NewEncoder(w).Encode(user)
}
