package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/rickydodd/media-rating-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var collection *mongo.Collection

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
}

func main() {
	r := gin.Default()

	r.GET("/media", listMedia)
	r.POST("/media", createMedia)
	r.GET("/media/:id", getMediaById)
	r.PUT("/media/:id", updateMediaRating)

	r.Run()
}

func listMedia(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	defer cur.Close(c)

	var medias []models.Media
	err = cur.All(ctx, &medias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, medias)
}

func createMedia(c *gin.Context) {
	var requestBody models.Media

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "malformed JSON in request body",
		})
		return
	}

	var media models.Media = requestBody
	media.ID = uuid.New()
	media.SubmittedRating = 0

	_, err := collection.InsertOne(ctx, media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

func getMediaById(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a UUID",
		})
		return
	}

	var media models.Media
	err = collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&media)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "media not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

func updateMediaRating(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	var requestBody models.Media

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a UUID",
		})
		return
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "malformed JSON in request body",
		})
		return
	}

	if requestBody.SubmittedRating > 10 || requestBody.SubmittedRating < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "mediaRating must be between 0 and 10, inclusive",
		})
		return
	}

	var media models.Media
	err = collection.FindOne(ctx, bson.M{
		"_id": id,
	}).Decode(&media)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "media not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	media.UnprocessedRating += requestBody.SubmittedRating
	media.NumberOfRatings += 1
	media.AverageRating = media.UnprocessedRating / float64(media.NumberOfRatings)

	_, err = collection.UpdateOne(ctx, bson.M{
		"_id": id,
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "ratingsCount", Value: media.NumberOfRatings},
		{Key: "unprocessedRating", Value: media.UnprocessedRating},
		{Key: "averageRating", Value: media.AverageRating},
	}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}
