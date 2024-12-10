package data

import "go.mongodb.org/mongo-driver/mongo"

type Repository interface {
	AllAudio() ([]*Audio, error)
	GetAudioByName(name string) (*Audio, error)
	GetAudioByID(id string) (*Audio, error)
	UpdateAudioByName(name, text string) (*Audio, error)
}

type mongoRepository struct {
	Collection *mongo.Collection
}

func newMongoRepository(collection *mongo.Collection) Repository {
	return &mongoRepository{Collection: collection}
}
