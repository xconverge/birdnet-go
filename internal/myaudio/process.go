// process.go
package myaudio

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tphakala/birdnet-go/internal/analysis/queue"
	"github.com/tphakala/birdnet-go/internal/birdnet"
	"github.com/tphakala/birdnet-go/internal/conf"
)

// processData processes the given audio data to detect bird species, logs the detected species
// and optionally saves the audio clip if a bird species is detected above the configured threshold.
func ProcessData(bn *birdnet.BirdNET, data []byte, startTime time.Time, source string) error {
	// get current time to track processing time
	predictStart := time.Now()

	// convert audio data to float32
	sampleData, err := ConvertToFloat32(data, conf.BitDepth)
	if err != nil {
		return fmt.Errorf("error converting %v bit PCM data to float32: %w", conf.BitDepth, err)
	}

	// run BirdNET inference
	results, err := bn.Predict(sampleData)
	if err != nil {
		return fmt.Errorf("error predicting species: %w", err)
	}

	// DEBUG print species of all results
	/*for i := 0; i < len(results); i++ {
		if results[i].Confidence > 0.01 {
			fmt.Println("	", results[i].Confidence, results[i].Species)
		}
	}*/

	// get elapsed time and log if enabled
	elapsedTime := logProcessingTime(predictStart)

	// Create a Results message to be sent through queue to processor
	resultsMessage := queue.Results{
		StartTime:   startTime,   // Timestamp when the audio data was received
		ElapsedTime: elapsedTime, // Time taken to process the audio data
		PCMdata:     data,        // BirdNET analyzed audio data
		Results:     results,     // Detected species and their confidence levels
		Source:      source,      // Source of the audio data, RSTP URL or audio card name
	}

	// Send the results to the queue
	select {
	case queue.ResultsQueue <- &resultsMessage:
		// Results enqueued successfully
	default:
		log.Println("Queue is full!")
		// Queue is full
	}
	return nil
}

func logProcessingTime(startTime time.Time) time.Duration {
	var elapsedTime = time.Since(startTime)
	/*if ctx.Settings.Realtime.ProcessingTime || ctx.Settings.Debug {
		fmt.Printf("\r\033[Kprocessing time %v ms", elapsedTime.Milliseconds())
		return elapsedTime
	}*/
	return elapsedTime
}

// ConvertToFloat32 converts a byte slice representing sample to a 2D slice of float32 samples.
// The function supports 16, 24, and 32 bit depths.
func ConvertToFloat32(sample []byte, bitDepth int) ([][]float32, error) {
	switch bitDepth {
	case 16:
		return [][]float32{convert16BitToFloat32(sample)}, nil
	case 24:
		return [][]float32{convert24BitToFloat32(sample)}, nil
	case 32:
		return [][]float32{convert32BitToFloat32(sample)}, nil
	default:
		return nil, errors.New("unsupported audio bit depth")
	}
}

// convert16BitToFloat32 converts 16-bit sample to float32 values.
func convert16BitToFloat32(sample []byte) []float32 {
	length := len(sample) / 2
	float32Data := make([]float32, length)
	divisor := float32(32768.0)

	for i := 0; i < length; i++ {
		sample := int16(sample[i*2]) | int16(sample[i*2+1])<<8
		float32Data[i] = float32(sample) / divisor
	}

	return float32Data
}

// convert24BitToFloat32 converts 24-bit sample to float32 values.
func convert24BitToFloat32(sample []byte) []float32 {
	length := len(sample) / 3
	float32Data := make([]float32, length)
	divisor := float32(8388608.0)

	for i := 0; i < length; i++ {
		sample := int32(sample[i*3]) | int32(sample[i*3+1])<<8 | int32(sample[i*3+2])<<16
		if (sample & 0x00800000) > 0 {
			sample |= ^0x00FFFFFF // Two's complement sign extension
		}
		float32Data[i] = float32(sample) / divisor
	}

	return float32Data
}

// convert32BitToFloat32 converts 32-bit sample to float32 values.
func convert32BitToFloat32(sample []byte) []float32 {
	length := len(sample) / 4
	float32Data := make([]float32, length)
	divisor := float32(2147483648.0)

	for i := 0; i < length; i++ {
		sample := int32(sample[i*4]) | int32(sample[i*4+1])<<8 | int32(sample[i*4+2])<<16 | int32(sample[i*4+3])<<24
		float32Data[i] = float32(sample) / divisor
	}

	return float32Data
}
