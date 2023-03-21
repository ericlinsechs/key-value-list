package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectionString = "mongodb://localhost:27017"
	dbName           = "mydb"
)

type Article struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

type Page struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Articles   []Article          `json:"articles"`
	NextPageID primitive.ObjectID `bson:"nextPageId,omitempty"`
}

type List struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	HeadPageID primitive.ObjectID `bson:"headpageId,omitempty"`
	Timestamp  int64              `json:"timestamp"`
}

const FirstPageKey = "000000000000000000000001"

func init() {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbName).Collection("pages")

	FirstPageID, err := primitive.ObjectIDFromHex(FirstPageKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Check if first page exists
	filter := bson.M{"_id": FirstPageID}
	page := Page{}
	err = collection.FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			FirstPage := Page{
				ID:         FirstPageID,
				Articles:   []Article{},
				NextPageID: primitive.NilObjectID,
			}
			result, err := collection.InsertOne(context.Background(), &FirstPage)
			if err != nil {
				log.Fatal(err)
			}
			InsertedID := result.InsertedID.(primitive.ObjectID)
			if InsertedID != FirstPageID {
				log.Fatal("InsertedID != FirstPageID")
			}
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/list/{list_id}", getHead).Methods("GET")
	r.HandleFunc("/page/getall", getAllPage).Methods("GET")
	r.HandleFunc("/page/get/{page_id}", getPage).Methods("GET")
	r.HandleFunc("/page/set/{page_id}", set).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func getHead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listID, err := primitive.ObjectIDFromHex(vars["list_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list := List{}
	head := Page{}

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbName).Collection("lists")

	filter := bson.M{"_id": listID}
	err = collection.FindOne(context.Background(), filter).Decode(&list)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.NotFound(w, r)
			return
		} else {
			log.Fatal(err)
		}
	}

	if list.HeadPageID.IsZero() {
		http.NotFound(w, r)
		return
	}

	collection = client.Database(dbName).Collection("pages")

	filter = bson.M{"_id": list.HeadPageID}
	err = collection.FindOne(context.Background(), filter).Decode(&head)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(head)
}

func getAllPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	var pages []Page

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbName).Collection("pages")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	err = cursor.All(ctx, &pages)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pages)

	defer cursor.Close(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID, err := primitive.ObjectIDFromHex(vars["page_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	page := Page{}

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbName).Collection("pages")

	filter := bson.M{"_id": pageID}
	err = collection.FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.NotFound(w, r)
			return
		} else {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

func set(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID, err := primitive.ObjectIDFromHex(vars["page_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("input pageID: %q\n", pageID)

	articles := []Article{}

	err = json.NewDecoder(r.Body).Decode(&articles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%q\n", articles)

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(dbName).Collection("pages")

	// Check if page exists
	filter := bson.M{"_id": pageID}
	page := Page{}
	err = collection.FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("create a new page.")
			nextPageID := primitive.NewObjectID()
			// Find the last inserted page and update its nextPageID field
			var lastPage Page
			filter := bson.M{"nextPageID": bson.M{"$exists": false}} //bson.M{"nextPageID": primitive.NilObjectID}
			err = collection.FindOne(context.Background(), filter).Decode(&lastPage)
			if err != nil {
				log.Println("Cannot find last inserted page.")
				log.Println("It is the first page.")
				// log.Fatal(err)
			} else {
				filter = bson.M{"_id": lastPage.ID}
				update := bson.M{"$set": bson.M{"nextPageID": nextPageID}}
				_, err = collection.UpdateOne(context.Background(), filter, update)
				if err != nil {
					log.Fatal(err)
				}
			}

			PageID := nextPageID
			// If page does not exist, create a new one
			nextPage := Page{
				ID:         PageID,
				Articles:   articles,
				NextPageID: primitive.NilObjectID,
			}

			log.Printf("NextPageID: %q\n", nextPage.ID)

			result, err := collection.InsertOne(context.Background(), &nextPage)
			if err != nil {
				log.Fatal(err)
			}
			InsertedID := result.InsertedID.(primitive.ObjectID)
			if InsertedID != PageID {
				log.Fatal("InsertedID != PageID")
			}
			log.Printf("InsertedID: %q\n", InsertedID)
		} else {
			log.Fatal(err)
		}
	} else {
		// If page exists, update its articles
		update := bson.M{"$set": bson.M{"articles": articles}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
