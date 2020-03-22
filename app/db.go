package app

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func openConnection() {

	db, err := sql.Open("mysql", config.DB_USER+":"+config.DB_PASS+"@tcp("+config.DB_HOST+":"+config.DB_PORT+")/"+config.DB_NAME+"?parseTime=true")

	if err != nil {
		log.Fatalf("Cannot open connection, %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot ping connectiong, %s", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	// The sum is the maximum number of concurrent connections
	db.SetConnMaxLifetime(5 * time.Minute)
}
