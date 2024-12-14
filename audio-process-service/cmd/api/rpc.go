package main

import (
	"audio_process/audio"
	"audio_process/data"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	mongodbUrl = "mongodb://localhost:27017"
)

type RPCServer struct {
	// App  *configuration.Application
	Repo *data.Models
}

type RPCPayload struct {
	Filename string `json:"filename"`
	Audio    string `json:"audio"` // Base64 encoded image
}

func (r *RPCServer) SaveAudio(payload RPCPayload, resp *string) error {

	// Decode the Base64 audio data
	audioData, err := base64.StdEncoding.DecodeString(payload.Audio)
	if err != nil {
		log.Println("error decoding audio file", err)
		return err
	}

	// Save the audio to disk
	audioPath := "./uploads/" + payload.Filename
	err = os.WriteFile(audioPath, audioData, 0644)
	if err != nil {
		log.Println("error saving to disk", err)
		return err
	}

	*resp = "Processed payload via RPC: " + payload.Filename

	r.ProcessAudio(audioPath, payload.Filename)

	return nil

}

func (r *RPCServer) ProcessAudio(filePath, filename string) {
	err := audio.ValidateFile(filePath)
	if err != nil {
		return
	}

	if !audio.IsValidWav(filePath) {
		filePath, err := audio.EncodeAudioToWav(filePath, "/uploads")
		SplitAudio(filePath, filename)
		if err != nil {
			return
		}
	} else {
		SplitAudio(filePath, filename)
	}

	client, err := initMongoDB(mongodbUrl)
	if err != nil {
		log.Panic(err)
		return
	}
	//r.App = configuration.New(client)
	r.Repo = data.New(client)

	audioText := GetTextFromAudio(filename)
	log.Println(filename, audioText)
	audio, err := r.Repo.Audio.UpdateAudioByName(filename, audioText)
	if err != nil {
		log.Panic(err)
		return
	}
	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	log.Println(audio)

}

func SplitAudio(inputFile, filename string) {
	// Split by "." and take the first part
	parts := strings.Split(filename, ".")
	folderName := parts[0]

	outputDir := "chunks/" + folderName // Directory to store audio chunks

	// Create output directory if it doesn't exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	// Get the total duration of the audio file
	totalDuration, err := audio.GetTotalDuration(inputFile)
	if err != nil {
		log.Fatalf("Failed to get total duration: %v", err)
	}
	fmt.Printf("Total Duration: %d seconds\n", totalDuration)

	chunkDuration := 30 // Chunk duration in seconds

	// Determine the number of chunks
	numChunks := (totalDuration + chunkDuration - 1) / chunkDuration

	// Set up worker pool
	numWorkers := 4 // Adjust based on your system capabilities
	results := make(chan audio.ChunkResult, numChunks)
	var wg sync.WaitGroup

	// Start processing chunks with a worker pool
	for i := 0; i < numChunks; i++ {
		startTime := i * chunkDuration
		wg.Add(1)
		go audio.ProcessChunk(inputFile, startTime, chunkDuration, outputDir, i+1, results, &wg)

		// Limit the number of concurrent workers
		if (i+1)%numWorkers == 0 {
			wg.Wait()
		}
	}

	// Wait for remaining workers to finish
	wg.Wait()
	close(results)

	// Collect results and ensure order
	var chunks []audio.ChunkResult
	for res := range results {
		if res.Err != nil {
			log.Printf("Error: %v\n", res.Err)
			continue
		}
		chunks = append(chunks, res)
	}

	// Sort chunks by order
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Order < chunks[j].Order
	})

	// Print the output files in order
	fmt.Println("Chunks processed in order:")
	for _, chunk := range chunks {
		fmt.Printf("Chunk %d: %s\n", chunk.Order, chunk.Path)
	}

}

func GetTextFromAudio(filename string) string {
	// Split by "." and take the first part
	parts := strings.Split(filename, ".")
	folderName := parts[0]
	outputDir := "./chunks/" + folderName // Directory to store audio chunks

	var audiosPath []string

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			audiosPath = append(audiosPath, path)
		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	var builder strings.Builder

	for _, audioFile := range audiosPath {

		text, err := audio.SpeechToText(audioFile)
		if err != nil {
			return ""
		}

		builder.WriteString(text)

	}
	audioText := builder.String()
	time.Sleep(60 * time.Second)
	DeleteFilesInFolder(outputDir)
	return audioText

}

// DeleteFilesInFolder deletes all files in the specified folder path
func DeleteFilesInFolder(folderPath string) error {
	// Read all files in the directory
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Iterate through the files and delete each one
	for _, file := range files {
		if !file.IsDir() { // Ensure it is a file, not a directory
			filePath := folderPath + string(os.PathSeparator) + file.Name()
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	return nil
}
