package app

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func openConnection() (*sql.DB, func() error) {

	db, err := sql.Open("mysql", config.DB_USER+":"+config.DB_PASS+"@tcp("+config.DB_HOST+":"+config.DB_PORT+")/"+config.DB_NAME)

	if err != nil {
		panic(err.Error())
	}
	return db, db.Close
}
