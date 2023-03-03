package models

import "github.com/google/uuid"

type Media struct {
	ID                uuid.UUID `json:"id" bson:"_id"`
	Title             string    `json:"mediaTitle" bson:"title"`
	ReleaseYear       string    `json:"mediaReleaseYear" bson:"releaseYear"`
	NumberOfRatings   uint64    `json:"-" bson:"ratingsCount"`
	SubmittedRating   float64   `json:"mediaRating,omitempty" bson:"-"`
	UnprocessedRating float64   `json:"-" bson:"unprocessedRating"`
	AverageRating     float64   `json:"mediaAverageRating" bson:"averageRating"`
}
