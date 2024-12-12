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

// UpdateAudioByName updates the audio document identified by its name
func (r *mongoRepository) UpdateAudioByName(name, text string) (*Audio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	// Find the audio document by name
	var audio Audio
	err := r.Collection.FindOne(ctx, bson.M{"audio_name": name}).Decode(&audio)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("document does not exist") // No document found with the specified name
		}
		return nil, err
	}

	// Update fields
	audio.UpdatedAt = time.Now()
	//audio.URL = "https://updated-url.com/audio.mp3" // Example: Update another field

	// Define the update query
	update := bson.M{
		"$set": bson.M{
			"audio_text": text,
			"updated_at": audio.UpdatedAt,
		},
	}

	// Perform the update
	result := r.Collection.FindOneAndUpdate(
		ctx,
		bson.M{"audio_name": name},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After), // Return the updated document
	)

	// Decode the updated document into an Audio struct
	var updatedAudio Audio
	err = result.Decode(&updatedAudio)
	if err != nil {
		return nil, err
	}

	return &updatedAudio, nil
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
