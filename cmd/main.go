package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/snappy"
	"log"
	"os"
	"runtime/debug"
	"sync"

	"ssh-plugin/config"
	"ssh-plugin/internal/discovery"
	"ssh-plugin/internal/metrics"
	"ssh-plugin/internal/models"
)

// main is the entry point for the plugin
// It reads JSON input from a file specified as a command-line argument,
// processes devices concurrently based on mode and system type,
// and streams JSON results to stdout as they arrive
// Panics are caught to prevent process crashes
func main() {

	log.SetOutput(os.Stderr)

	// Recover from panics in main
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Fatal panic: %v, stack: %s\n", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	// Check command-line arguments
	if len(os.Args) != 3 {
		log.Printf("Usage: %s <mode> <file_path>\n", os.Args[0])
		os.Exit(1)
	}

	mode := os.Args[1]
	filePath := os.Args[2]

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Read devices from file
	devices, err := decryptAndDecompressFile(filePath, cfg)
	if err != nil {
		log.Printf("Error reading devices: %v\n", err)
		os.Exit(1)
	}

	// Validate input
	if len(devices) == 0 {
		log.Printf("No devices provided in input\n")
		os.Exit(1)
	}

	// Process devices and stream results
	switch mode {
	case "metrics":
		processMetrics(devices, cfg)
	case "discovery":
		processDiscovery(devices, cfg)
	default:
		log.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}

	// Exit with success (0) unless a critical failure occurred
	os.Exit(0)
}

// readDevicesFromFile reads devices from a file, handling compression and encryption
func decryptAndDecompressFile(filePath string, cfg *config.Config) ([]models.Device, error) {

	// Step 0: check if the key exists in config
	keyHex := cfg.Encryption.Key
	if keyHex == "" {
		return nil, fmt.Errorf("encryption key not found in config")
	}

	// Step 1: Read Base64-encoded content
	base64Content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Step 2: Decode Base64
	decodedBytes, err := base64.StdEncoding.DecodeString(string(base64Content))
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %w", err)
	}

	if len(decodedBytes) < 12 {
		return nil, fmt.Errorf("data too short: missing nonce")
	}

	// Step 3: Extract nonce and ciphertext
	nonce := decodedBytes[:12]
	ciphertext := decodedBytes[12:]

	// Step 4: Decode AES key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}

	// Step 5: AES-GCM decryption
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES cipher creation failed: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCM mode failed: %w", err)
	}

	compressed, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	// Step 6: Snappy decompress
	decompressed, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("snappy decompress failed: %w", err)
	}

	// Step 7: Parse JSON
	var devices []models.Device
	if err := json.Unmarshal(decompressed, &devices); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return devices, nil
}

// processMetrics processes devices concurrently for metrics collection,
// dispatching based on system type and streaming results to stdout
func processMetrics(devices []models.Device, cfg *config.Config) {

	key, err := hex.DecodeString(cfg.Encryption.Key)
	if err != nil {
		log.Printf("Invalid encryption key: %v\n", err)
		return
	}

	// Channel to receive results
	resultChan := make(chan models.MetricsResult, len(devices))

	var wg sync.WaitGroup
	var outputWg sync.WaitGroup // WaitGroup for the output Goroutine

	// Start a Goroutine to stream results as JSON
	outputWg.Add(1)
	go func() {
		// Ensure the output Goroutine signals completion
		defer outputWg.Done()
		// Recover from panics in the output Goroutine
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in output goroutine: %v, stack: %s\n", r, string(debug.Stack()))
			}
		}()

		for result := range resultChan {
			encoded, err := encodeResult(result, key)
			if err != nil {
				log.Printf("Error encoding result for device %d: %v\n", result.ID, err)
				continue
			}
			fmt.Println(string(encoded))
		}
	}()

	// Process each device in a Goroutine
	for _, device := range devices {
		wg.Add(1)
		go func(dev models.Device) {
			defer wg.Done()
			// Recover from panics in the processing Goroutine
			defer func() {
				if r := recover(); r != nil {
					resultChan <- models.NewMetricsError(dev.ID, fmt.Sprintf("panic recovered: %v, stack: %s", r, string(debug.Stack())))
				}
			}()

			// Dispatch based on system type
			collector := metrics.GetMetricsCollector(dev.SystemType)
			result := collector.Collect(dev, cfg.SSH.Timeout)
			resultChan <- result
		}(device)
	}

	// Wait for all device-processing Goroutines to complete
	wg.Wait()

	close(resultChan)
	// Wait for the output Goroutine to finish printing
	outputWg.Wait()
}

// processDiscovery processes devices concurrently for SSH discovery,
// dispatching based on system type and streaming results to stdout
func processDiscovery(devices []models.Device, cfg *config.Config) {

	// Channel to receive results
	resultChan := make(chan models.DiscoveryResult, len(devices))

	var wg sync.WaitGroup
	var outputWg sync.WaitGroup // WaitGroup for the output Goroutine

	// Start a Goroutine to stream results as JSON
	outputWg.Add(1)
	go func() {
		// Ensure the output Goroutine signals completion
		defer outputWg.Done()
		// Recover from panics in the output Goroutine
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in output goroutine: %v, stack: %s\n", r, string(debug.Stack()))
			}
		}()

		for result := range resultChan {
			// Marshal the result to JSON
			output, err := json.Marshal(result)
			if err != nil {
				log.Printf("Error encoding result for device %d: %v\n", result.ID, err)
				continue
			}
			fmt.Println(string(output))
		}
	}()

	// Process each device in a Goroutine
	for _, device := range devices {
		wg.Add(1)
		go func(dev models.Device) {
			defer wg.Done()
			// Recover from panics in the processing Goroutine
			defer func() {
				if r := recover(); r != nil {
					resultChan <- models.NewDiscoveryResult(dev.ID, false, fmt.Sprintf("panic recovered: %v, stack: %s", r, string(debug.Stack())))
				}
			}()

			// Dispatch based on system type
			performer := discovery.GetDiscoveryPerformer(dev.SystemType)
			result := performer.Perform(dev, cfg.SSH.Timeout)
			resultChan <- result
		}(device)
	}

	// Wait for all device-processing Goroutines to complete
	wg.Wait()

	close(resultChan)
	// Wait for the output Goroutine to finish printing
	outputWg.Wait()
}

func encodeResult(result models.MetricsResult, key []byte) (string, error) {
	plaintext, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	compressed := snappy.Encode(nil, plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher error: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm error: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("nonce error: %w", err)
	}

	encrypted := gcm.Seal(nil, nonce, compressed, nil)

	final := append(nonce, encrypted...) // prepend nonce
	encoded := base64.StdEncoding.EncodeToString(final)
	return encoded, nil
}
