package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
)

// Create and Retrieve Users.

func (c *appContext) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User
	c.db.Select(&users, "SELECT * FROM users ORDER BY id ASC")

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(&UsersResource{Data: users})
}

func (c *appContext) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var user User
	c.db.Get(&user, "SELECT * FROM users WHERE username=$1", params.ByName("id"))

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(&UserResource{Data: user})
}

func (c *appContext) createUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*UserResource)
	user := body.Data

	// Check if this username is already taken.
	users := []User{}
	c.db.Select(&users, "SELECT * FROM users WHERE username=$1", user.Username)
	if len(users) != 0 {
		WriteError(w, ErrUserExists)
		return
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	var newUser User
	c.db.Get(&newUser, "SELECT * FROM users WHERE username=$1", user.Username)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(&UserResource{Data: newUser})
}

// Followers

const userfollowers = `
SELECT users.id, users.username FROM users 
JOIN user_followers 
ON user_followers.follower_id=users.id 
WHERE user_followers.user_id=$1
`

func (c *appContext) getUserFollowersHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []User

	log.Println("userId: ", userId)

	c.db.Select(&users, userfollowers, userId)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(UsersResource{Data: users})
}

const usersFollowing = `
SELECT users.id, users.username FROM users 
JOIN user_followers 
ON user_followers.user_id=users.id 
WHERE user_followers.follower_id=$1
`

func (c *appContext) getUsersFollowingHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("x-key")
	var users []User

	log.Println("userId: ", userId)

	c.db.Select(&users, usersFollowing, userId)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(UsersResource{Data: users})
}

func (c *appContext) followUserHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	var userToFollow User
	var user User

	userToFollowId := params.ByName("id")
	userId := r.Header.Get("x-key")

	c.db.Get(&userToFollow, "SELECT * FROM users WHERE id=$1", userToFollowId)
	c.db.Get(&user, "SELECT * FROM users WHERE id=$1", userId)

	if userToFollow.Id == 0 || user.Id == 0 {
		WriteError(w, ErrBadRequest)
		return
	}

	tx := c.db.MustBegin()
	tx.MustExec("INSERT INTO user_followers (user_id, follower_id) VALUES ($1, $2)", userToFollow.Id, user.Id)
	if tx.Commit() != nil {
		WriteError(w, ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(user)
}
