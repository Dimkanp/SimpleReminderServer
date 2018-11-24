package main

import (
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type AuthData struct {
	Login    string
	Password string
}

type User struct {
	Id            int64
	Login         string
	PasswordHash  string
	Email         string
	Surname       string
	Name          string
	LastLogin     int64
	Notifications []Notification
}

type Notification struct {
	Id               int64
	UnixSelectedDate int64
	ReminderText     string
	UserId           int
}

func main() {
	http.HandleFunc("/authorise", authorizeUser)
	http.HandleFunc("/register", addUser)
	http.HandleFunc("/user_notifications", getUserNotifications)
	http.HandleFunc("/add_notification", addNotification)
	http.HandleFunc("/remove_notification", deleteNotification)
	http.HandleFunc("/update_notification", editNotification)

	isExit := make(chan bool, 1)
	go func() {
		fmt.Println("Server started successfully.")
		fmt.Println(http.ListenAndServe(":8080", nil))
		isExit <- true
	}()
	<-isExit
	closeDatabaseConnection()
	fmt.Println("Server stopped!")
}
