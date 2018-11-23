package main

import (
	"fmt"
	"net/http"
	"time"

	//"github.com/mattn/go-sqlite3"
)

type User struct {
	Login string
	Password string
	Email string
}

type Notification struct {
	Id int
	NotificationTime time.Time
	Text string
}

func addUser(w http.ResponseWriter, r *http.Request) {
}

func addNotification(w http.ResponseWriter, r *http.Request)  {

}

func deleteNotification(w http.ResponseWriter, r *http.Request)  {

}

func editNotification(w http.ResponseWriter, r *http.Request)  {

}

func main() {
	http.HandleFunc("/register", addUser)
	http.HandleFunc("/add_notification", addNotification)
	fmt.Println(http.ListenAndServe(":8080", nil))
}