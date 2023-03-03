package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rickydodd/media-rating-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MediaHandler struct {
	ctx        context.Context
	collection *mongo.Collection
}

func NewMediaHandler(ctx context.Context, collection *mongo.Collection) *MediaHandler {
	return &MediaHandler{
		ctx:        ctx,
		collection: collection,
	}
}

func (handler MediaHandler) ListMedia(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	defer cur.Close(c)

	var medias []models.Media
	err = cur.All(handler.ctx, &medias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, medias)
}

func (handler MediaHandler) CreateMedia(c *gin.Context) {
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

	_, err := handler.collection.InsertOne(handler.ctx, media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

func (handler MediaHandler) GetMediaById(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a UUID",
		})
		return
	}

	var media models.Media
	err = handler.collection.FindOne(handler.ctx, bson.M{
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

func (handler MediaHandler) UpdateMediaRating(c *gin.Context) {
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
	err = handler.collection.FindOne(handler.ctx, bson.M{
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

	_, err = handler.collection.UpdateOne(handler.ctx, bson.M{
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
