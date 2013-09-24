package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func Db(dbConfig DbConfig) *sql.DB {
	fmt.Println(dbConfig.User, dbConfig.Host)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbConfig.User,
		dbConfig.Pass, dbConfig.Host, dbConfig.Port, dbConfig.Name))
	if err != nil {
		log.Fatal("Unable to connect to the database: ", err)
	}
	return db
}

func SetClicks(db *sql.DB, Clicks []Click) {
	if len(Clicks) < 1 {
		return
	}
	stmt, err := db.Prepare("INSERT INTO click(url, ip_address, href, user_agent) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Println("Unable to prepare statment for Click: ", err)
	}
	for k := range Clicks {
		stmt.Exec(
			Clicks[k].URL,
			Clicks[k].IPAddress,
			Clicks[k].Href,
			Clicks[k].UserAgent,
		)
	}
}
