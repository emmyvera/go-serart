package audio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type AudioFlie struct {
	Name         string
	FileLocation string
}

// isValidWav checks if a .wav file is valid using ffprobe
func isValidWav(filePath string) (bool, error) {
	cmd := exec.Command("ffprobe", "-i", filePath, "-show_streams", "-select_streams", "a", "-loglevel", "error")
	err := cmd.Run()
	if err != nil {
		return false, errors.New("invalid .wav file")
	}
	return true, nil
}

// convertToWav converts an audio/video file to .wav using ffmpeg
func EncodeAudioToWav(inputFile, outputDir string) (string, error) {
	// Get the file name without the extension
	fileName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	outputFile := filepath.Join(outputDir, fileName+".wav")

	// Use ffmpeg to convert the file
	cmd := exec.Command("ffmpeg", "-i", inputFile,
		"-acodec", "pcm_s16le",
		"-ar", "44100", outputFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.New("failed to convert file to .wav")
	}

	return outputFile, nil
}

// validateFile checks if the input file is a valid audio or video file
func ValidateFile(inputFile string) error {
	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

	// Check file extension
	validExtensions := []string{".mp3", ".mp4", ".wav", ".aac", ".flac", ".ogg", ".mkv", ".mov"}
	ext := strings.ToLower(filepath.Ext(inputFile))
	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil
		}
	}
	return errors.New("unsupported file format")
}

// Metadata represents the structure to parse FFprobe JSON output
type Metadata struct {
	Streams []struct {
		Duration string `json:"duration"`
	} `json:"streams"`
}

// getTotalDuration retrieves the total duration of an audio file in seconds
func getTotalDuration(inputFile string) (int, error) {
	cmd := exec.Command("ffprobe", "-i", inputFile, "-show_streams", "-select_streams", "a", "-print_format", "json", "-loglevel", "error")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("failed to get file metadata: %v", err)
	}

	// Parse the JSON output
	var metadata Metadata
	err = json.Unmarshal(out.Bytes(), &metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to parse metadata: %v", err)
	}

	// Ensure we have at least one audio stream
	if len(metadata.Streams) == 0 || metadata.Streams[0].Duration == "" {
		return 0, errors.New("no audio stream found in the file")
	}

	// Convert duration to an integer
	duration, err := strconv.ParseFloat(metadata.Streams[0].Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return int(duration), nil
}

// ChunkResult stores the result of a chunk with its order
type ChunkResult struct {
	Order int
	Path  string
	Err   error
}

// Worker function to process audio chunks
func processChunk(inputFile string, startTime int, duration int, outputDir string, order int, results chan<- ChunkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	outputFile := filepath.Join(outputDir, fmt.Sprintf("chunk_%03d.wav", order))
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-ss", strconv.Itoa(startTime), "-t", strconv.Itoa(duration), outputFile, "-y")

	err := cmd.Run()
	if err != nil {
		results <- ChunkResult{Order: order, Path: "", Err: fmt.Errorf("failed to process chunk %d: %v", order, err)}
		return
	}

	results <- ChunkResult{Order: order, Path: outputFile, Err: nil}
}
