package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var database * sql.DB
var dbName = "simple-reminder-database.db"

func authorizeUser(responseWriter http.ResponseWriter, request *http.Request) {
	var authData AuthData
	err := readBodyJson(request.Body, &authData)
	if err != nil {
		fmt.Println("authorizeUser | Parsing input data error:", err)
		badRequest(responseWriter, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println("authorizeUser | Getting database connection error: ", err)
		internalError(responseWriter,"Can't connect to database.")
		return
	}

	userSelectSqlStmt := `
	SELECT id,
	       login,
		   password,
		   email,
		   surname,
		   name,
		   lastLogin
    FROM "User"
	WHERE login = '%s' AND password = '%s';`

	var user User
	tmp := fmt.Sprintf(userSelectSqlStmt,
					   authData.Login,
					   authData.Password)
	row := db.QueryRow(tmp)
	if row == nil {
		fmt.Println("authorizeUser | User (login: '",authData.Login,"', password: *** doesn't exist.")
		badRequest(responseWriter,"Invalid pair of login and password.")
		return
	}
	err = row.Scan(
			&user.Id,
			&user.Login,
			&user.PasswordHash,
			&user.Email,
			&user.Surname,
			&user.Name,
			&user.LastLogin)
	if err != nil {
		fmt.Println("authorizeUser | Scanning QueryRow result error: ", err)
		internalError(responseWriter,"Can't execute sql query.")
		return
	}
	user.Notifications, _ = getNotificationsByUserId(user.Id)

	err = writeBodyJson(responseWriter,user)
	if err != nil{
		fmt.Println("authorizeUser | Sending response error: ", err)
	}
}

func addUser(responseWriter http.ResponseWriter, request *http.Request) {
	var user User
	err := readBodyJson(request.Body, &user)
	if err != nil {
		fmt.Println("addUser | Parsing input data error: ", err)
		badRequest(responseWriter, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println("addUser | Getting database connection error: ", err)
		internalError(responseWriter,"Can't connect to database.")
		return
	}

	//Checking login for use by someone else
	userIdSelectSqlStmt := `
	SELECT id 
    FROM "User" 
    WHERE login = '%s';`

	id := 0
	row := db.QueryRow(fmt.Sprintf(userIdSelectSqlStmt, user.Login))
	if row.Scan(&id) == nil {
		fmt.Println("addUser | User login '", user.Login ,"' already exist.")
		badRequest(responseWriter,"User login already exist.")
		return
	}

	userAddSqlStmt := `
	INSERT INTO "User" (login,
						password,
						email,
						surname,
						name,
						lastLogin) 
	VALUES ('%s', '%s', '%s', '%s', '%s', %d);`
	tmp := fmt.Sprintf(userAddSqlStmt,
		user.Login,
		user.PasswordHash,
		user.Email,
		user.Surname,
		user.Name,
		user.LastLogin)
	result, _ := db.Exec(tmp)
	user.Id, err = result.LastInsertId()
	if err != nil {
		fmt.Println("addUser | Executing sql statement error: ", err, "\n	SQL: '", tmp, "'")
		internalError(responseWriter,"Can't execute sql query.")
		return
	}
	err = writeBodyJson(responseWriter,fmt.Sprint(user.Id))
	if err != nil{
		fmt.Println("addUser | Sending response error: ", err)
	}
}

func getUserNotifications(responseWriter http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("userId")
	userId, err := strconv.ParseInt(id, 10, 8)
	if err != nil {
		fmt.Println("getUserNotifications | Parsing input data error: ", err)
		badRequest(responseWriter, "Can't parse data.")
		return
	}

	notifications, err := getNotificationsByUserId(userId)
	if err != nil{
		internalError(responseWriter, "Can't get notifications from DataBase.")
		return
	}

	err = writeBodyJson(responseWriter,notifications)
	if err != nil{
		fmt.Println("getUserNotifications | Sending response error: ", err)
	}
}

func addNotification(responseWriter http.ResponseWriter, request *http.Request) {
	var notification Notification
	err := readBodyJson(request.Body, &notification)
	if err != nil {
		fmt.Println("addNotification | Parsing input data error: ", err)
		badRequest(responseWriter, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println("addNotification | Getting database connection error: ", err)
		internalError(responseWriter,"Can't connect to database.")
		return
	}

	notificationAddSqlStmt := `
	INSERT INTO "Notification" (unixSelectedDate,
								reminderText,
								userId) 
	VALUES (%d, '%s', %d);`
	tmp := fmt.Sprintf(notificationAddSqlStmt,
		notification.UnixSelectedDate,
		notification.ReminderText,
		notification.UserId)
	row, _ := db.Exec(tmp)
	notification.Id, err = row.LastInsertId()
	if err != nil {
		fmt.Println("addNotification | Executing sql statement error: ", err, "\n	SQL: '", tmp, "'")
		internalError(responseWriter,"Can't execute sql query.")
		return
	}
	err = writeBodyJson(responseWriter,fmt.Sprint(notification.Id))
	if err != nil{
		fmt.Println("addNotification | Sending response error: ", err)
	}
}

func deleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("notificationId"))
	if err != nil {
		fmt.Println("deleteNotification | Parsing input data error: ", err)
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println("deleteNotification | Getting database connection error: ", err)
		internalError(w,"Can't connect to database.")
		return
	}

	notificationDeleteSqlStmt := `
	DELETE FROM "Notification"
	WHERE id = %d;`
	var rowsAffected int64
	tmp := fmt.Sprintf(notificationDeleteSqlStmt, id)
	row, _ := db.Exec(tmp)
	rowsAffected, err = row.RowsAffected()
	if err != nil || rowsAffected == 0 {
		fmt.Println("deleteNotification | Executing sql statement error: ", err, "\n	SQL: '", tmp, "'")
		internalError(w,"Can't execute sql query.")
		return
	}

	writeBody(w,"")
}

func editNotification(w http.ResponseWriter, r *http.Request) {
	var notification Notification
	err := readBodyJson(r.Body, &notification)
	if err != nil {
		fmt.Println("editNotification | Parsing input data error: ", err)
		badRequest(w, "Can't parse data.")
		return
	}

	db, err := getDB()
	if err != nil {
		fmt.Println("editNotification | Getting database connection error: ", err)
		internalError(w,"Can't connect to database.")
		return
	}

	notificationEditSqlStmt := `
	UPDATE "Notification"
	SET unixSelectedDate = %d,
		reminderText = '%s',
		userId = %d
	WHERE id = %d;`
	var rowsAffected int64
	tmp := fmt.Sprintf(notificationEditSqlStmt,
		notification.UnixSelectedDate,
		notification.ReminderText,
		notification.UserId,
		notification.Id)
	row, _ := db.Exec(tmp)
	rowsAffected, err = row.RowsAffected()
	if err != nil || rowsAffected == 0 {
		fmt.Println("editNotification | Executing sql statement error: ", err, "\n	SQL: '", tmp, "'")
		internalError(w,"Can't execute sql query.")
		return
	}

	writeBody(w,"")
}

func getNotificationsByUserId(userId int64) ([]Notification, error) {
	notifications := []Notification{}
	db, err := getDB()
	if err != nil {
		fmt.Println("getNotificationsByUserId | Getting database connection error: ", err)
		return notifications, err
	}

	notificationsSelectSqlStmt := `
	SELECT id, 
           unixSelectedDate,
           reminderText,
           userId
	FROM "Notification" 
	WHERE userId = ?;`

	rows, err := db.Query(notificationsSelectSqlStmt, userId)
	defer rows.Close()
	for rows.Next(){
		notification := new(Notification)
		err = rows.Scan(
			&notification.Id,
			&notification.UnixSelectedDate,
			&notification.ReminderText,
			&notification.UserId)
		if err != nil {
			fmt.Println("getNotificationsByUserId | Scanning notification error: ", err)
		} else {
			notifications = append(notifications, *notification)
		}
	}
	if err != nil {
		fmt.Println("getNotificationsByUserId | Getting database connection error: ", err)
		return notifications, err
	}
	return notifications, err
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
			fmt.Println("Open connection with database error: ", err)
			return nil, err
		}
	}
	return database, nil
}

func initDatabase(dbName string){
	fmt.Println("Creating a new database file.")
	var err error
	database, err = sql.Open("sqlite3", fmt.Sprint("./", dbName))
	if err != nil {
		fmt.Println("Creating database file error: ", err)
		return
	}

	creatingUserTableSqlStmt := `
	CREATE TABLE "User" (id INTEGER NOT NULL PRIMARY KEY, 
						 login TEXT,
						 password TEXT,
						 email TEXT,
						 surname TEXT,
						 name TEXT,
						 lastLogin INTEGER);
	DELETE FROM "User";`
	_, err = database.Exec(creatingUserTableSqlStmt)
	if err != nil {
		fmt.Println("Executing sql statement error: ", err, "\n	SQL: '", creatingUserTableSqlStmt, "'")
		deleteDatabaseFile(dbName)
		return
	}

	creatingNotificationTableSqlStmt := `
	CREATE TABLE "Notification" (id INTEGER NOT NULL PRIMARY KEY,
								 unixSelectedDate INTEGER,
								 reminderText TEXT,
								 userId INTEGER,
								 FOREIGN KEY(userId) REFERENCES "User"(id));
	DELETE FROM "Notification";`
	_, err = database.Exec(creatingNotificationTableSqlStmt)
	if err != nil {
		fmt.Println("Executing sql statement error: ", err, "\n	SQL: '", creatingNotificationTableSqlStmt, "'")
		deleteDatabaseFile(dbName)
		return
	}
	fmt.Println("New database file successfully created.")
}

func deleteDatabaseFile(dbName string){
	fmt.Println("Deleting database file '", dbName, "'")
	var err = os.Remove(fmt.Sprint("./", dbName))
	if err != nil {
		fmt.Println("Deleting database file error: ", err)
		return
	}
	fmt.Println("Database file has been deleted.")
}

func closeDatabaseConnection(){
	if database != nil{
		database.Close()
	}
}
