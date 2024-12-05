package data

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var repo Repository

type Models struct {
	Audio Audio
}

func New(client *mongo.Client) *Models {

	if client != nil {
		collection := client.Database("serart").Collection("audio")

		repo = newMongoRepository(collection)
	} else {
		repo = newTestRepository(nil)
	}

	return &Models{
		Audio: Audio{},
	}

}

type Audio struct {
	ID        int       `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	URL       string    `bson:"url" json:"url"`
	Text      string    `bson:"text,omitempty" json:"text,omitempty"`
	AudioName string    `bson:"audio_name" json:"audio_name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (a *Audio) AllAudio() ([]*Audio, error) {
	return repo.AllAudio()
}

func (a *Audio) GetAudioByName(n string) (*Audio, error) {
	return repo.GetAudioByName(n)
}

func (a *Audio) AddAudio(audio *Audio) (*mongo.InsertOneResult, error) {
	return repo.AddAudio(audio)
}
