package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var db* sql.DB
var dbName = "simple-reminder-database.db"

func authorizeUser(w http.ResponseWriter, r *http.Request) {

}

func addUser(w http.ResponseWriter, r *http.Request) {

}

func getUserNotifications(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 8)
	fmt.Println(userId)
	// TODO return user notification using his ID
	// Remove fmt and write code with data base
}

func addNotification(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	readBodyJson(r.Body, &notification)
	fmt.Println(notification)
}

func deleteNotification(w http.ResponseWriter, r *http.Request) {

}

func editNotification(w http.ResponseWriter, r *http.Request) {

}

func getDB() (*sql.DB, error){
	if _, err := os.Stat(fmt.Sprint("./", dbName)); os.IsNotExist(err){
		initDatabase(dbName)
	}

	var err error
	db, err = sql.Open("sqlite3", fmt.Sprint("./", dbName))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func initDatabase(dbName string){
	var err error
	db, err = sql.Open("sqlite3", fmt.Sprint("./", dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userSqlStmt := `
	create table "User" (id integer not null primary key, 
						 login text,
						 password text,
						 email text,
						 surname text,
						 name text,
						 lastLogin integer);
	delete from "User";`

	notificationSqlStmt := `
	create table "Notification" (id integer not null primary key,
								 unixSelectedDate integer,
								 reminderText text,
								 userId integer);
	delete from "Notification";`

	_, err = db.Exec(userSqlStmt)
	_, err = db.Exec(notificationSqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, userSqlStmt)
		return
	}
}

func closeDatabaseConnection(){
	if db != nil{
		db.Close()
	}
}
