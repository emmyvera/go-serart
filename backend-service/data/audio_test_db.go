package data

import "go.mongodb.org/mongo-driver/mongo"

func (r *testRepository) AllAudio() ([]*Audio, error) {
	return nil, nil
}

func (r *testRepository) GetAudioByName(name string) (*Audio, error) {
	return nil, nil
}

func (r *testRepository) GetAudioByID(id string) (*Audio, error) {

	return nil, nil
}

func (r *testRepository) AddAudio(audio *Audio) (*mongo.InsertOneResult, error) {
	return nil, nil
}

func (r *testRepository) DeleteAudioByID(id string) error {
	return nil
}
