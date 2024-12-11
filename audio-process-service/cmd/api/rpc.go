package main

import (
	"audio_process/audio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type RPCServer struct{}

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
	err = os.WriteFile("./uploads/"+payload.Filename, audioData, 0644)
	if err != nil {
		log.Println("error saving to disk", err)
		return err
	}

	*resp = "Processed payload via RPC: " + payload.Filename

	return nil

}

func ProcessAudio(filePath, fileName string) {
	err := audio.ValidateFile(filePath)
	if err != nil {
		return
	}

	if !audio.IsValidWav(filePath) {
		filePath, err := audio.EncodeAudioToWav(fileName, "/uploads")
		SplitAudio(filePath)
		if err != nil {
			return
		}
	} else {
		SplitAudio(filePath)
	}

}

func SplitAudio(inputFile string) {
	//inputFile := "example.wav" // Replace with your WAV file path
	outputDir := "chunks" // Directory to store audio chunks

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

func SaveTextToDB(name string) {
	var audiosPath []string

	err := filepath.Walk("./uploads/chunk", func(path string, info os.FileInfo, err error) error {
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
			return
		}

		builder.WriteString(text)

	}

	//audioText := builder.String()

}
