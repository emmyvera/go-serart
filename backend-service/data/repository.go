package data

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	AllAudio() ([]*Audio, error)
	GetAudioByName(name string) (*Audio, error)
	GetAudioByID(id string) (*Audio, error)
	AddAudio(audio *Audio) (*mongo.InsertOneResult, error)
	DeleteAudioByID(id string) error
}

// mongoRepository is the implementation of the Repository interface for MongoDB
type mongoRepository struct {
	Collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository
func newMongoRepository(collection *mongo.Collection) Repository {
	return &mongoRepository{Collection: collection}
}

type testRepository struct {
	Collection *mongo.Collection
}

func newTestRepository(collection *mongo.Collection) Repository {
	return &testRepository{
		Collection: collection,
	}
}
