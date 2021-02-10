package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

var (
	client *mongo.Client
	err    error
	ctx    context.Context
)

func main() {
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	router := mux.NewRouter()
	router.HandleFunc("/posts", GetPosts).Methods("GET")
	router.HandleFunc("/add", InsertPost).Methods("POST")
	router.HandleFunc("/{id}", GetPost).Methods("GET")
	router.HandleFunc("/{id}", DeletePost).Methods("DELETE")
	router.HandleFunc("/update", UpdatePost).Methods("PUT")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func InsertPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	json.NewDecoder(r.Body).Decode(&post)
	collection := client.Database("lupin").Collection("post")
	result, err := collection.InsertOne(ctx, post)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(result)
}

func GetPosts(w http.ResponseWriter, _ *http.Request) {
	collection := client.Database("lupin").Collection("post")
	var posts []Post
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(ctx) {
		var post Post
		err = cur.Decode(&post)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, post)
	}
	_ = json.NewEncoder(w).Encode(posts)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	var post Post
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("lupin").Collection("post")
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(post)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("lupin").Collection("post")
	res, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(w).Encode(res)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	json.NewDecoder(r.Body).Decode(&post)
	collection := client.Database("lupin").Collection("post")
	_, err := collection.UpdateOne(ctx, bson.M{"_id": post.Id},
		bson.D{
			{"$set", bson.D{{"title", post.Title}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(post)
}

type Post struct {
	Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title string             `json:"title,omitempty"`
	Body  string             `json:"body,omitempty"`
}
