package app

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func openConnection() (*sql.DB, func() error) {

	db, err := sql.Open("mysql", config.DB_USER+":"+config.DB_PASS+"@tcp("+config.DB_HOST+":"+config.DB_PORT+")/"+config.DB_NAME+"?parseTime=true")

	if err != nil {
		log.Printf(err.Error())
		return nil, nil
	}

	err = db.Ping()
	if err != nil {
		log.Printf(err.Error())
		return nil, nil
	}

	return db, db.Close
}
