package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AllAudio retrieves all audio documents
func (r *mongoRepository) AllAudio() ([]*Audio, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.Collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var audios []*Audio
	for cursor.Next(ctx) {
		var audio Audio
		if err := cursor.Decode(&audio); err != nil {
			return nil, err
		}
		audios = append(audios, &audio)
	}
	return audios, nil
}

// GetAudioByName retrieves an audio document by name
func (r *mongoRepository) GetAudioByName(name string) (*Audio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var audio Audio
	err := r.Collection.FindOne(ctx, bson.M{"name": name}).Decode(&audio)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &audio, nil
}

// GetAudioByID retrieves an audio document by ID
func (r *mongoRepository) GetAudioByID(id string) (*Audio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var audio Audio
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&audio)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &audio, nil
}

// AddAudio inserts a new audio document
func (r *mongoRepository) AddAudio(audio *Audio) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return r.Collection.InsertOne(ctx, audio)
}

// DeleteAudioByID deletes an audio document by ID
func (r *mongoRepository) DeleteAudioByID(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
