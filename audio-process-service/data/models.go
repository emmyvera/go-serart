package data

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var repo Repository

type Models struct {
	Audio Audio
}

type Audio struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	URL       string    `bson:"url" json:"url"`
	AudioText string    `bson:"audio_text," json:"audio_text"`
	AudioName string    `bson:"audio_name" json:"audio_name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func New(client *mongo.Client) *Models {
	if client != nil {
		collection := client.Database("serart").Collection("audio")
		repo = newMongoRepository(collection)
		log.Println("Client not null")
	}

	return &Models{
		Audio: Audio{},
	}
}

func (a *Audio) AllAudio() ([]*Audio, error) {
	return repo.AllAudio()
}

func (a *Audio) UpdateAudioByName(name, text string) (*Audio, error) {
	return repo.UpdateAudioByName(name, text)
}

func (a *Audio) GetAudioByName(n string) (*Audio, error) {
	return repo.GetAudioByName(n)
}
