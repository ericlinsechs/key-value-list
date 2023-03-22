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
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

const (
	connectionString = "mongodb://localhost:27017"
	dbName           = "shared-key-value-list"
)

const (
	FirstListKey = "000000000000000000000001"
	FirstPageKey = "000000000000000000000100"
)

type Article struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `json:"title"`
	Author  string             `json:"author"`
	Content string             `json:"content"`
}

type Page struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	ListID     primitive.ObjectID   `bson:"list_id,omitempty"`
	ArticleID  []primitive.ObjectID `bson:"article_id"`
	NextPageID primitive.ObjectID   `bson:"nextPageID,omitempty"`
}

type List struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	HeadPageID primitive.ObjectID `bson:"headpageId,omitempty"`
}

var client *mongo.Client
var listsCollection *mongo.Collection
var pagesCollection *mongo.Collection
var articlesCollection *mongo.Collection

func init() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	listsCollection = client.Database(dbName).Collection("lists", options.Collection().SetReadConcern(readconcern.Majority()))
	pagesCollection = client.Database(dbName).Collection("pages", options.Collection().SetReadConcern(readconcern.Majority()))
	articlesCollection = client.Database(dbName).Collection("articles", options.Collection().SetReadConcern(readconcern.Majority()))

	list := List{}

	FirstListID, err := primitive.ObjectIDFromHex(FirstListKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	FirstPageID, err := primitive.ObjectIDFromHex(FirstPageKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Check if first list exists
	filter := bson.M{"_id": FirstListID}
	err = listsCollection.FindOne(context.Background(), filter).Decode(&list)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create list if it doesn't be found in database
			FirstList := List{
				ID:         FirstListID,
				HeadPageID: FirstPageID,
			}
			result, err := listsCollection.InsertOne(context.Background(), &FirstList)
			if err != nil {
				log.Fatal(err)
			}
			InsertedID := result.InsertedID.(primitive.ObjectID)
			if InsertedID != FirstListID {
				log.Fatal("InsertedID != FirstListID")
			}
		}
	} else {
		log.Printf("list %q already exist.", FirstListID)
	}

	// Check if first page exists
	filter = bson.M{"_id": FirstPageID}
	page := Page{}
	err = pagesCollection.FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			FirstPage := Page{
				ID:         FirstPageID,
				ListID:     FirstListID,
				NextPageID: primitive.NilObjectID,
			}
			result, err := pagesCollection.InsertOne(context.Background(), &FirstPage)
			if err != nil {
				log.Fatal(err)
			}
			InsertedID := result.InsertedID.(primitive.ObjectID)
			if InsertedID != FirstPageID {
				log.Fatal("InsertedID != FirstPageID")
			}
		}
	} else {
		log.Printf("page %q already exist.", FirstPageID)
	}
}

func main() {
	defer client.Disconnect(context.Background())
	r := mux.NewRouter()

	// list
	r.HandleFunc("/list/{list_id}", getHead).Methods("GET")

	// page
	r.HandleFunc("/page/getall", getAllPage).Methods("GET")
	r.HandleFunc("/page/get/{page_id}", getPage).Methods("GET")
	r.HandleFunc("/page/set/{page_id}", set).Methods("POST")

	// article
	r.HandleFunc("/article/getall", getAllArticle).Methods("GET")
	r.HandleFunc("/article/create", createArticle).Methods("POST")

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

	filter := bson.M{"_id": listID}
	err = listsCollection.FindOne(context.Background(), filter).Decode(&list)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list.HeadPageID)
}

func getAllPage(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	var pages []Page

	cursor, err := pagesCollection.Find(ctx, bson.M{})
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
	err = json.NewEncoder(w).Encode(pages)
	if err != nil {
		log.Fatal(err)
	}
}

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID, err := primitive.ObjectIDFromHex(vars["page_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	page := Page{}

	filter := bson.M{"_id": pageID}
	err = pagesCollection.FindOne(context.Background(), filter).Decode(&page)
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
	//***************************************************************************************************
	// Set the pipeline for the query
	pipeline := bson.A{
		bson.M{"$sample": bson.M{"size": 5}},
		bson.M{"$project": bson.M{"_id": 0, "title": 1, "author": 1, "content": 1}},
	}

	// Execute the query and get the results
	cursor, err := articlesCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		// Handle the error
	}

	defer cursor.Close(context.Background())

	var articles []Article
	if err := cursor.All(context.Background(), &articles); err != nil {
		// Handle the error
	}

	// Check if page exists
	filter := bson.M{"_id": pageID}
	page := Page{}
	err = pagesCollection.FindOne(context.Background(), filter).Decode(&page)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			nextPageID := primitive.NewObjectID()
			// Find the last inserted page and update its nextPageID field
			var lastPage Page
			filter := bson.M{"nextPageID": bson.M{"$exists": false}}
			// filter := bson.M{"nextPageID": primitive.NilObjectID}
			err = pagesCollection.FindOne(context.Background(), filter).Decode(&lastPage)
			if err != nil {
				log.Println("Cannot find the last inserted page.")
				log.Fatal(err)
			} else {
				filter = bson.M{"_id": lastPage.ID}
				update := bson.M{"$set": bson.M{"nextPageID": nextPageID}}
				result, err := pagesCollection.UpdateOne(context.Background(), filter, update)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("matched: %v, modified: %v\n", result.MatchedCount, result.ModifiedCount)
			}

			PageID := nextPageID
			// If page does not exist, create a new one
			nextPage := Page{
				ID:         PageID,
				ListID:     lastPage.ListID,
				NextPageID: primitive.NilObjectID,
			}

			result, err := pagesCollection.InsertOne(context.Background(), &nextPage)
			if err != nil {
				log.Fatal(err)
			}
			InsertedID := result.InsertedID.(primitive.ObjectID)
			if InsertedID != PageID {
				log.Fatal("InsertedID != PageID")
			}
		} else {
			log.Fatal(err)
		}
	} else {
		// If page exists, update its articles
		// update := bson.M{"$set": bson.M{"article_id": articles}}
		// _, err = pagesCollection.UpdateOne(context.Background(), filter, update)
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAllArticle(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	var articles []Article

	cursor, err := articlesCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	err = cursor.All(ctx, &articles)
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(articles)

	defer cursor.Close(ctx)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(articles)
	if err != nil {
		log.Fatal(err)
	}
}

func createArticle(w http.ResponseWriter, r *http.Request) {
	var articles []Article

	err := json.NewDecoder(r.Body).Decode(&articles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	articlesToInsert := make([]interface{}, len(articles))
	for i, article := range articles {
		articlesToInsert[i] = article
	}
	res, err := articlesCollection.InsertMany(context.Background(), articlesToInsert)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("inserted documents with IDs %v\n", res.InsertedIDs)
}
