package main

// importing required packages

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// schema for user

type MongoUserSchema struct {
	ID       primitive.ObjectID `json: "_id,omitempty"`
	Name     string             `json: "name"`
	Email    string             `json: "email"`
	Password string             `json: "password"`
}

// schema for post

type MongoPostSchema struct {
	ID        primitive.ObjectID `json: "_id,omitempty"`
	Caption   string             `json: "caption"`
	ImageURL  string             `json: "imageURL"`
	Timestamp time.Time          `json: "timestamp"`
}

// declaring mongo client variable globally

var client *mongo.Client

// function to check whether entered email is valid

func checkEmailValidity(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// function to cryptographically hash a password using SHA256

func hashPassword(pswd string) string {
	data := []byte(pswd)
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// handler function for post route: to create a new user

func createUserHandler(res http.ResponseWriter, req *http.Request) {
	var user MongoUserSchema
	json.NewDecoder(req.Body).Decode(&user)
	// fmt.Println(hash(user.Password))
	if checkEmailValidity(user.Email) == false {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Invalid e-mail id!"))
		return
	}

	usersCol := client.Database("Aviroop_Nandy_Appointy").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	cursor, err := usersCol.Find(ctx, bson.M{})

	for cursor.Next(ctx) {
		var backlogUser MongoUserSchema
		cursor.Decode(&backlogUser)
		if backlogUser.Email == user.Email {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(`{"This e-mail is already registered!":"` + err.Error() + `"}`))
			return
		}
	}

	hashedPswd := hashPassword(user.Password)
	user.Password = hashedPswd

	userResult, insertErrorUser := usersCol.InsertOne(ctx, user)
	if insertErrorUser != nil {
		fmt.Println("Error while creating user: ", insertErrorUser)
	} else {
		json.NewEncoder(res).Encode(userResult)
		userID := userResult.InsertedID
		fmt.Println("New user id: ", userID)
	}

	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
}

// handler function for get route: to get details of a user by passing id in URL parameter

func getUserHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var user MongoUserSchema
	usersCol := client.Database("Aviroop_Nandy_Appointy").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err := usersCol.FindOne(ctx, MongoUserSchema{ID: id}).Decode(&user)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"Error message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(user)
}

// handler function for post route: to create a new post

func createPostHandler(res http.ResponseWriter, req *http.Request) {
	var post MongoPostSchema
	post.Timestamp = time.Now()
	json.NewDecoder(req.Body).Decode(&post)

	postsCol := client.Database("Aviroop_Nandy_Appointy").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	postResult, insertErrorPost := postsCol.InsertOne(ctx, post)

	if insertErrorPost != nil {
		fmt.Println("Error while creating post: ", insertErrorPost)
	} else {
		newPostID := postResult.InsertedID
		fmt.Println("New post ID: ", newPostID)
	}

	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
}

// handler function for get route: to get details of a post by passing id in URL parameter

func getPostHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application.json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post MongoPostSchema

	postsCol := client.Database("Aviroop_Nandy_Appointy").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err := postsCol.FindOne(ctx, MongoPostSchema{ID: id}).Decode(&post)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"Error message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(post)
}

// handler function for get route: to get details of all posts of a user by passing id in the URL parameter

func getUserPostsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "application/json")
	params := mux.Vars(req)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println(id)
	var posts []MongoPostSchema
	postsCol := client.Database("Aviroop_Nandy_Appointy").Collection("posts")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := postsCol.Find(ctx, bson.M{})

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"Error message":"` + err.Error() + `"}`))
		return
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post MongoPostSchema
		cursor.Decode(&post)
		posts = append(posts, post)
	}

	if err := cursor.Err(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(res).Encode(posts)
}

func main() {
	fmt.Println("Main.go up and running!")

	// establishing connection with MongoDB

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	client, _ = mongo.Connect(ctx, clientOptions)

	// listing all routes

	http.HandleFunc("/users", createUserHandler)
	http.HandleFunc("/users/{id}", getUserHandler)
	http.HandleFunc("/posts", createPostHandler)
	http.HandleFunc("/posts/{id}", getPostHandler)
	http.HandleFunc("/posts/users/{id}", getUserPostsHandler)

	// server listening on port 8000

	httpErr := http.ListenAndServe(":8000", nil)

	// server error handler

	if httpErr != nil {
		panic(httpErr)
	}
}
