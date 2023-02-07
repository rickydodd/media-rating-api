package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Media struct {
	Id                uuid.UUID `json:"mediaId"`
	Title             string    `json:"mediaTitle"`
	ReleaseYear       string    `json:"mediaReleaseYear"`
	NumberOfRatings   uint64    `json:"-"`
	SubmittedRating   float64   `json:"mediaRating,omitempty"`
	UnprocessedRating float64   `json:"-"`
	AverageRating     float64   `json:"mediaAverageRating,omitempty"`
}

var Medias map[uuid.UUID]Media

func main() {
	r := gin.Default()

	r.GET("/media", func(c *gin.Context) {
		c.JSON(http.StatusOK, Medias)
	})
	r.POST("/media", createMedia)
	r.GET("/media/:id", getMediaById)
	r.PUT("/media/:id", updateMediaRating)

	r.Run()
}

func init() {
	Medias = make(map[uuid.UUID]Media)
}

func createMedia(c *gin.Context) {
	var requestBody Media

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed JSON in request body",
		})
		return
	}

	requestBody.Id = uuid.New()
	requestBody.SubmittedRating = 0
	Medias[requestBody.Id] = requestBody

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully created media",
	})
}

func getMediaById(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "mediaId is not a UUID",
		})
		return
	}

	media := Medias[id]

	if media == *new(Media) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Media not found",
		})
		return
	}

	c.JSON(http.StatusOK, Medias[id])
}

func updateMediaRating(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	var requestBody Media

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "mediaId is not a UUID",
		})
		return
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Malformed JSON in request body",
		})
		return
	}

	if requestBody.SubmittedRating > 10 || requestBody.SubmittedRating < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "mediaRating must be between 0 and 10, inclusive",
		})
		return
	}

	media := Medias[id]

	if media == *new(Media) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Media not found",
		})
	}

	media.UnprocessedRating += requestBody.SubmittedRating
	media.NumberOfRatings += 1
	media.AverageRating = media.UnprocessedRating / float64(media.NumberOfRatings)
	Medias[id] = media
	c.JSON(http.StatusOK, Medias[id])
}
