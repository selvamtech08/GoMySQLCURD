package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// User struct
type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

// response for successfully operation
type infoResponse struct {
	Message string `json:"info"`
}

// response for failed operation
type errorResponse struct {
	Message string `json:"error"`
}

// insert new user in db
func insertUserdb(newUser *User) error {
	query := "INSERT INTO users (name, email, location) VALUES (?, ?, ?)"
	_, err := db.Exec(query, newUser.Name, newUser.Email, newUser.Location)
	if err != nil {
		log.Println("db insert: ", err.Error())
		return errors.New("failed to add the user")
	}
	return nil
}

// get all users from db
func getAllUsers() (*[]User, error) {
	query := "SELECT * FROM users"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	var users []User
	for rows.Next() {
		var user User
		var userid int // just ignored show to user
		if err := rows.Scan(&userid, &user.Name, &user.Email, &user.Location); err != nil {
			// for debug
			log.Println("db get: failed to scan the values from row object,\n\t" + err.Error())
			return nil, errors.New("failed to scan the values from db")
		} else {
			users = append(users, user)
		}
	}
	return &users, nil
}

// get a user from db using given name
func getaUserdb(name string) (*User, error) {
	query := "SELECT * FROM users WHERE name = ?"
	row := db.QueryRow(query, name)
	var user User
	var userId int
	if err := row.Scan(&userId, &user.Name, &user.Email, &user.Location); err != nil {
		log.Println("db get: failed to parse the db row,\n\t" + err.Error())
		return nil, errors.New("invalid username, failed to get user information from db")
	}
	return &user, nil
}

// update user in db
func updateUserdb(user *User) error {
	query := "UPDATE users SET name = ?, email = ?, location = ? where name = ?"
	_, err := db.Exec(query, user.Name, user.Email, user.Location, user.Name)
	if err != nil {
		log.Println("db update: ", err.Error())
		return errors.New("failed to update given user details")
	}
	return nil
}

// remove user from db
func removeUserdb(userName string) error {
	query := "DELETE FROM users WHERE name = ?"
	row, err := db.Exec(query, userName)
	if err != nil {
		log.Println("db remove: ", err.Error())
		return errors.New("failed to remove the user")
	}
	resp, _ := row.RowsAffected()
	if resp == 0 {
		log.Println("db remove: 0 row affected")
		return fmt.Errorf("given user `%s` not exists", userName)
	}
	return nil
}

// show all the users from db
func getAllUsersHandler(c *gin.Context) {
	users, err := getAllUsers()
	if err != nil {
		log.Println("failed to collect user informations from db" + err.Error())
		c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "failed to collect user informations from db",
		})
	}
	c.JSON(http.StatusOK, users)
}

// show the required user info
func getUserHandler(c *gin.Context) {
	userName, ok := c.Params.Get("username")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			Message: "failed to prase the request",
		})
		return
	}
	user, err := getaUserdb(userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			err.Error(),
		})
		return
	}
	c.JSON(http.StatusAccepted, user)
}

// add new user
func addNewUserHandler(c *gin.Context) {
	// init new user object
	var newUser User

	// parse json data from request body and assing to newuser object
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Message: "failed to parse the request, " + err.Error(),
		})
		return
	}
	log.Println(newUser)
	if err := insertUserdb(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "failed to update user details in db, please contact admin",
		})
		return
	}
	c.JSON(http.StatusAccepted, infoResponse{
		Message: fmt.Sprintf("new user `%s` added successfully!", newUser.Name),
	})
}

// update the user details
func updateUserHandler(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		log.Println("update handler", err)
		c.JSON(http.StatusBadRequest, errorResponse{
			Message: "faliled to prase the request, verify the given details and try again",
		})
		return
	}
	if err := updateUserdb(&user); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusAccepted, infoResponse{
		Message: fmt.Sprintf("given user `%s` detail has been updated", user.Name),
	})
}

// remove the user
func removeUserHandler(c *gin.Context) {
	userName, ok := c.Params.Get("username")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			Message: "invalid request",
		})
		return
	}
	if err := removeUserdb(userName); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusAccepted, infoResponse{
		Message: fmt.Sprintf("given user `%s` removed", userName),
	})

}

func main() {

	// define base route
	router := gin.Default()

	// connect nysql db
	var err error
	db, err = sql.Open("mysql", "username:password@tcp(localhost:port)/dbname")
	if err != nil {
		log.Panicln("failed to connect db", err.Error())
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Panicln("failed to pind the db connection", err.Error())
	}

	// add middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// routes
	router.GET("/", getAllUsersHandler)
	router.GET("/:username", getUserHandler)
	router.POST("/", addNewUserHandler)
	router.PUT("/", updateUserHandler)
	router.DELETE("/:username", removeUserHandler)

	// start the service
	log.Fatal(router.Run("localhost:8081"))
}
