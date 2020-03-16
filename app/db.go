package app

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func openConnection() (*sql.DB, func() error, error) {

	db, err := sql.Open("mysql", config.DB_USER+":"+config.DB_PASS+"@tcp("+config.DB_HOST+":"+config.DB_PORT+")/"+config.DB_NAME+"?parseTime=true")

	if err != nil {
		return nil, nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, nil, err
	}

	return db, db.Close, err
}
