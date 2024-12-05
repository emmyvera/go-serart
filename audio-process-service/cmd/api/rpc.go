package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
)

type RPCServer struct{}

type RPCPayload struct {
	Filename string `json:"filename"`
	Audio    string `json:"audio"` // Base64 encoded image
}

func (r *RPCServer) SaveAudio(payload RPCPayload, resp *string) error {

	// err := json.NewDecoder().Decode(&payload)
	//  if err != nil {
	// 	log.Println("error decoding json payload", err)
	//  	return err
	//  }

	// Decode the Base64 audio data
	audioData, err := base64.StdEncoding.DecodeString(payload.Audio)
	if err != nil {
		log.Println("error decoding audio file", err)
		return err
	}

	// Save the audio to disk
	err = ioutil.WriteFile("./uploads/"+payload.Filename, audioData, 0644)
	if err != nil {
		log.Println("error saving to disk", err)
		return err
	}

	*resp = "Processed payload via RPC: " + payload.Filename

	return nil

}
