package main

import (
	"fmt"
	"net/http"
	"strconv"
)

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
