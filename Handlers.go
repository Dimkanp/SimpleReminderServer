package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var database * sql.DB
var dbName = "simple-reminder-database.db"

func authorizeUser(w http.ResponseWriter, r *http.Request) {
	var authData AuthData
	err := readBodyJson(r.Body, &authData)
	if err != nil {
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		log.Fatal(err)
		internalError(w,"Can't connect to database.")
		return
	}

	userSelectSqlStmt := `
	select id,
	       login,
		   password,
		   email,
		   surname,
		   name,
		   lastLogin
    from "User"
	where login = '%s' and password = '%s';`

	var user User
	tmp := fmt.Sprintf(userSelectSqlStmt,
					   authData.Login,
					   authData.Password)
	err = db.QueryRow(tmp).Scan(
			&user.Id,
			&user.Login,
			&user.PasswordHash,
			&user.Email,
			&user.Surname,
			&user.Name,
			&user.LastLogin)
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't execute sql query.")
		return
	}
	user.Notifications, _ = getNotificationsByUserId(user.Id)

	//fmt.Println(user)

	writeBodyJson(w,user)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := readBodyJson(r.Body, &user)
	if err != nil {
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		log.Fatal(err)
		internalError(w,"Can't connect to database.")
		return
	}

	//Checking login for use by someone else
	stmt, err := db.Prepare(`select id from "User" where login = ?;`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(user.Login).Scan(&user.Id)
	if err == nil {
		badRequest(w,"User login already exist.")
		return
	}

	userAddSqlStmt := `
	insert into "User" (login,
						password,
						email,
						surname,
						name,
						lastLogin) 
	values ('%s', '%s', '%s', '%s', '%s', %d);`
	tmp := fmt.Sprintf(userAddSqlStmt,
		user.Login,
		user.PasswordHash,
		user.Email,
		user.Surname,
		user.Name,
		user.LastLogin)
	row, _ := db.Exec(tmp)
	user.Id, err = row.LastInsertId()
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't execute sql query.")
		return
	}
	writeBodyJson(w,fmt.Sprint(user.Id))
}

func getUserNotifications(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("userId")
	userId, err := strconv.ParseInt(id, 10, 8)
	if err != nil {
		badRequest(w, "Can't parse data.")
		return
	}

	notifications, _ := getNotificationsByUserId(userId)

	writeBodyJson(w,notifications)
}

func addNotification(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	err := readBodyJson(r.Body, &notification)
	if err != nil {
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't connect to database.")
		return
	}

	userAddSqlStmt := `
	insert into "Notification" (unixSelectedDate,
								reminderText,
								userId) 
	values (%d, '%s', %d);`
	tmp := fmt.Sprintf(userAddSqlStmt,
		notification.UnixSelectedDate,
		notification.ReminderText,
		notification.UserId)
	row, _ := db.Exec(tmp)
	notification.Id, err = row.LastInsertId()
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't execute sql query.")
		return
	}
	writeBodyJson(w,fmt.Sprint(notification.Id))
}

func deleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("notificationId"))
	if err != nil {
		badRequest(w, err.Error())
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't connect to database.")
		return
	}

	sqlStatement := `
	delete from "Notification"
	where id = %d;`
	var rowsAffected int64
	tmp := fmt.Sprintf(sqlStatement, id)
	row, _ := db.Exec(tmp)
	rowsAffected, err = row.RowsAffected()
	if err != nil || rowsAffected == 0 {
		fmt.Println(err)
		internalError(w,"Can't execute sql query.")
		return
	}

	writeBody(w,"")
}

func editNotification(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	err := readBodyJson(r.Body, &notification)
	if err != nil {
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println(err)
		internalError(w,"Can't connect to database.")
		return
	}

	sqlStatement := `
	update "Notification"
	set unixSelectedDate = %d,
		reminderText = '%s',
		userId = %d
	where id = %d;`
	var rowsAffected int64
	tmp := fmt.Sprintf(sqlStatement,
		notification.UnixSelectedDate,
		notification.ReminderText,
		notification.UserId,
		notification.Id)
	row, _ := db.Exec(tmp)
	rowsAffected, err = row.RowsAffected()
	if err != nil || rowsAffected == 0 {
		fmt.Println(err)
		internalError(w,"Can't execute sql query.")
		return
	}

	writeBody(w,"")
}

func getDB() (*sql.DB, error) {
	if _, err := os.Stat(fmt.Sprint("./", dbName)); os.IsNotExist(err){
		fmt.Println("Database file ",dbName," not found.")
		initDatabase(dbName)
	}

	if database == nil {
		var err error
		database, err = sql.Open("sqlite3", fmt.Sprint("./", dbName))
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	return database, nil
}

func initDatabase(dbName string){
	var err error
	database, err = sql.Open("sqlite3", fmt.Sprint("./", dbName))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Creating a new database file.")

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

	_, err = database.Exec(userSqlStmt)
	_, err = database.Exec(notificationSqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, userSqlStmt)
		return
	}
	fmt.Println("New database file successfully created.")
}

func closeDatabaseConnection(){
	if database != nil{
		database.Close()
	}
}

func getNotificationsByUserId(userId int64) ([]Notification, error) {
	db, err := getDB()
	notifications := []Notification{}
	if err != nil {
		fmt.Println(err)
		return notifications, err
	}

	notificationsSelectSqlStatement := `
	select id, 
           unixSelectedDate,
           reminderText,
           userId
	from "Notification" 
	where userId = ?;`

	rows, err := db.Query(notificationsSelectSqlStatement, userId)
	defer rows.Close()
	for rows.Next(){
		notification := new(Notification)
		err = rows.Scan(
			&notification.Id,
			&notification.UnixSelectedDate,
			&notification.ReminderText,
			&notification.UserId)
		if err != nil {
			fmt.Println(err)
		}
		notifications = append(notifications, *notification)
	}
	if err != nil {
		fmt.Println(err)
		return notifications, err
	}
	return notifications, err
}
