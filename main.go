package main

import (
	"context"
	"log"
	"os"

	"github.com/rickydodd/media-rating-api/handlers"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var collection *mongo.Collection
var mediaHandler *handlers.MediaHandler

func init() {
	ctx = context.Background()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("media")
	mediaHandler = handlers.NewMediaHandler(ctx, collection)
}

func main() {
	r := gin.Default()

	r.GET("/media", mediaHandler.ListMedia)
	r.POST("/media", mediaHandler.CreateMedia)
	r.GET("/media/:id", mediaHandler.GetMediaById)
	r.PUT("/media/:id", mediaHandler.UpdateMediaRating)

	r.Run()
}
