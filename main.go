package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type MyUrl struct {
	ID       string `json:"id,omitempty" bson:"id,omitempty"`
	LongUrl  string `json:"longurl,omitempty" bson:"longurl,omitempty"`
	ShortUrl string `json:"shorturl,omitempty" bson:"shorturl,omitempty"`
}

var client *mongo.Client

// TODO:
// ExpandEndpoint is for getting the original url from the
// shortened one.
// GET /expand/
// func ExpandEndpoint(w http.ResponseWriter, r *http.Request) {

// }

func CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var url MyUrl
	json.NewDecoder(r.Body).Decode(&url)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	urlDatabase := client.Database("urls")
	urlCollection := urlDatabase.Collection("urldata")
	_, err := urlCollection.InsertOne(ctx, url)
	if err != nil {
		log.Fatal(err)
	}

	var row MyUrl
	if err := urlCollection.FindOne(ctx, bson.M{}).Decode(&row); err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	hd := hashids.NewData()
	h, _ := hashids.NewWithData(hd)
	now := time.Now()
	url.ID, _ = h.Encode([]int{int(now.Unix())})
	url.ShortUrl = "http://localhost:12345/" + url.ID
	_, err = urlCollection.InsertOne(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(url)

}

func RootEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var url MyUrl
	blogCollection := client.Database("urls").Collection("urldata")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	blogCollection.FindOne(ctx, bson.M{"id": id}).Decode(&url)
	http.Redirect(w, r, url.LongUrl, http.StatusFound)
}

func main() {
	r := mux.NewRouter()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://dbUser:secret-password@cluster0.ehlxz.mongodb.net/admin?retryWrites=true&w=majority"))

	r.HandleFunc("/create", CreateEndpoint).Methods("POST")
	// r.HandleFunc("/expand/", ExpandEndpoint).Methods("GET")
	r.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	fmt.Println("Starting server on port 12345...")
	log.Fatal(http.ListenAndServe(":12345", r))
}
