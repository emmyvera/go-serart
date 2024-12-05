package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"serart_be/data"
	"serart_be/event"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tsawler/toolbox"
)

const (
	MaxFileSize = 80 * 1024 * 1024 // 80MB
)

type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type RPCPayload struct {
	Filename string `json:"filename"`
	Audio    string `json:"audio"` // Base64 encoded image
}

type RMQPayload struct {
	Name string     `json:"name"`
	Data RPCPayload `json:"data"`
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

func (app *Config) Greet(w http.ResponseWriter, r *http.Request) {
	var t toolbox.Tools

	var payload JsonResponse
	payload.Error = false
	payload.Message = "Hello World"

	t.WriteJSON(w, http.StatusOK, payload)
}

func (app *Config) Audio(w http.ResponseWriter, r *http.Request) {
	file, header, _ := r.FormFile("file")
	defer file.Close()

	// create a destination file
	dst, err := os.Create(filepath.Join("./", header.Filename))
	if err != nil {
		log.Panicf("error creating destination folder %v", err)
		return
	}
	defer dst.Close()

	// upload the file to destination path
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Panicf("error storing file in folder")
		return
	}

	fmt.Println("File uploaded successfully")
}

func (app *Config) SendAudio(w http.ResponseWriter, r *http.Request) {
	var t toolbox.Tools

	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)

	// Retrieve name data
	name := r.FormValue("name")

	// Parse the multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		t.ErrorJSON(w, errors.New("file too large"), http.StatusRequestEntityTooLarge)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		t.ErrorJSON(w, errors.New("failed to read file"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uniqueFilename := generateUniqueFilename(fileHeader.Filename)

	// // Connect to Cloudinary
	// cld, err := cloudinary.NewFromParams("doo0g2su7", "163285136689312", "av5FENVBcCPzyZrvP_MxjNLWknk")
	// if err != nil {
	// 	t.ErrorJSON(w, errors.New("error initializing Cloudinary"), http.StatusInternalServerError)
	// 	return
	// }

	// // Upload file to Cloudinary
	// uploadResult, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
	// 	Folder:   "audio_files",
	// 	PublicID: uniqueFilename, // Optional: specify a name
	// })
	// if err != nil {
	// 	t.ErrorJSON(w, errors.New("error uploading to Cloudinary"), http.StatusInternalServerError)
	// 	return
	// }

	// Save the file to the server
	dst, err := os.Create(filepath.Join("./uploads", uniqueFilename))
	if err != nil {
		t.ErrorJSON(w, errors.New("error saving file"), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		t.ErrorJSON(w, errors.New("error copying file"), http.StatusInternalServerError)
		return
	}

	// Save the audio metadata to MongoDB
	audio := Audio{
		Name:      name,
		URL:       uniqueFilename, // TODO change to cloudinary url
		Text:      r.FormValue("text"),
		AudioName: uniqueFilename,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	app.App.Models.Audio.AddAudio((*data.Audio)(&audio))

	// Encode file to Base64
	//filePath := "./uploads" + uniqueFilename
	audioData, err := os.ReadFile(filepath.Join("./uploads", uniqueFilename))
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	base64Audio := base64.StdEncoding.EncodeToString(audioData)

	t.WriteJSON(w, http.StatusAccepted, audio)
	rpcPayload := RPCPayload{
		Filename: uniqueFilename,
		Audio:    base64Audio,
	}
	rqPayload := RMQPayload{
		Name: "process",
		Data: rpcPayload,
	}
	app.processAudioViaRabbit(rqPayload)

}

// Generate a unique filename with the original extension
func generateUniqueFilename(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName)) // Get file extension
	if ext == "" {
		ext = ".bin" // Default extension if none provided
	}
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}

func (app *Config) processAudioViaRabbit(p RMQPayload) {
	err := app.pushToQueue(p.Name, p.Data)
	if err != nil {
		log.Print(err)
		return
	}

	// var payload jsonResponse
	// payload.Error = false
	// payload.Message = "logged via RabbitMQ"

	log.Print("Successfully sent")

}

func (app *Config) pushToQueue(name string, data RPCPayload) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := RMQPayload{
		Name: name,
		Data: data,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")

	if err != nil {
		return err
	}

	return nil
}
