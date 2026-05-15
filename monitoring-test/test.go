package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type TestConfig struct {
	BaseURL    string
	NumWorkers int
	NumRequests int
}

type LoadTestResult struct {
	TotalRequests int64
	Successful    int64
	Failed        int64
	EndpointStats map[string]int64
}

func RunSimpleLoadTest(config TestConfig) *LoadTestResult {
	fmt.Printf("Starting load test on %s\n", config.BaseURL)
	fmt.Printf("Workers: %d, Total requests: %d\n", config.NumWorkers, config.NumRequests)

	var totalRequests int64
	var successful int64
	var failed int64
	var endpointStats = make(map[string]int64)
	var statsMu sync.Mutex

	var wg sync.WaitGroup
	requestsPerWorker := config.NumRequests / config.NumWorkers

	// Start workers
	for w := 0; w < config.NumWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			client := &http.Client{Timeout: 10 * time.Second}
			
			for i := 0; i < requestsPerWorker; i++ {
				// Randomly choose an endpoint
				endpoints := []string{"/notes", "/whoami", "/healthz", "/metrics"}
				endpoint := endpoints[rand.Intn(len(endpoints))]
				
				var resp *http.Response
				var err error
				start := time.Now()
				
				// Make request
				switch endpoint {
				case "/notes":
					if rand.Intn(2) == 0 {
						// GET /notes
						resp, err = client.Get(config.BaseURL + "/notes")
						trackStats("GET /notes", resp, err, &totalRequests, &successful, &failed, &statsMu, endpointStats)
					} else {
						// POST /notes
						note := map[string]string{
							"title":   fmt.Sprintf("Test note %d", rand.Intn(10000)),
							"content": "Test content",
						}
						jsonData, _ := json.Marshal(note)
						resp, err = client.Post(config.BaseURL+"/notes", "application/json", bytes.NewBuffer(jsonData))
						trackStats("POST /notes", resp, err, &totalRequests, &successful, &failed, &statsMu, endpointStats)
					}
					
				case "/whoami":
					resp, err = client.Get(config.BaseURL + "/whoami")
					trackStats("GET /whoami", resp, err, &totalRequests, &successful, &failed, &statsMu, endpointStats)
					
				case "/healthz":
					resp, err = client.Get(config.BaseURL + "/healthz")
					trackStats("GET /healthz", resp, err, &totalRequests, &successful, &failed, &statsMu, endpointStats)
					
				case "/metrics":
					resp, err = client.Get(config.BaseURL + "/metrics")
					trackStats("GET /metrics", resp, err, &totalRequests, &successful, &failed, &statsMu, endpointStats)
				}
				
				// Ensure response body is closed
				if resp != nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				
				// Calculate latency
				latency := time.Since(start)
				if latency > 1*time.Second {
					fmt.Printf("Slow request to %s took %v\n", endpoint, latency)
				}
				
				// Small delay between requests
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}(w)
	}

	wg.Wait()

	return &LoadTestResult{
		TotalRequests: atomic.LoadInt64(&totalRequests),
		Successful:    atomic.LoadInt64(&successful),
		Failed:        atomic.LoadInt64(&failed),
		EndpointStats: endpointStats,
	}
}
func trackStats(endpoint string, resp *http.Response, err error, 
	total, success, failed *int64, statsMu *sync.Mutex, stats map[string]int64) {
	
	atomic.AddInt64(total, 1)
	
	if err != nil {
		atomic.AddInt64(failed, 1)
		statsMu.Lock()
		stats[fmt.Sprintf("%s (error: %v)", endpoint, err)]++
		statsMu.Unlock()
		return
	}
	
	// Добавьте искусственные задержки для разных эндпоинтов
	if resp != nil {
		switch {
		case resp.StatusCode == 200:
			// Быстрые запросы - нет задержки
		case resp.StatusCode == 201:
			// Создание ресурса - небольшая задержка
			time.Sleep(50 * time.Millisecond)
		case resp.StatusCode == 404:
			// Not found - средняя задержка
			time.Sleep(100 * time.Millisecond)
		case resp.StatusCode >= 500:
			// Ошибки сервера - большая задержка
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		atomic.AddInt64(success, 1)
	} else {
		atomic.AddInt64(failed, 1)
	}
	
	statsMu.Lock()
	stats[fmt.Sprintf("%s (status: %d)", endpoint, resp.StatusCode)]++
	statsMu.Unlock()
}

func CreateTestNotes(baseURL string, count int) []string {
	client := &http.Client{Timeout: 5 * time.Second}
	var ids []string
	
	fmt.Printf("Creating %d test notes...\n", count)
	
	for i := 0; i < count; i++ {
		note := map[string]string{
			"title":   fmt.Sprintf("Test Note %d", i+1),
			"content": fmt.Sprintf("This is test note %d content", i+1),
		}
		
		jsonData, err := json.Marshal(note)
		if err != nil {
			fmt.Printf("Failed to marshal note %d: %v\n", i, err)
			continue
		}
		
		resp, err := client.Post(baseURL+"/notes", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create note %d: %v\n", i, err)
			continue
		}
		
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var result map[string]interface{}
			body, _ := io.ReadAll(resp.Body)
			json.Unmarshal(body, &result)
			if id, ok := result["id"].(string); ok {
				ids = append(ids, id)
				fmt.Printf("✓ Created note %d with ID: %s\n", i+1, id)
			}
		}
		resp.Body.Close()
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("Successfully created %d/%d test notes\n", len(ids), count)
	return ids
}

func DeleteTestNotes(baseURL string, ids []string) {
	if len(ids) == 0 {
		return
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	fmt.Printf("Deleting %d test notes...\n", len(ids))
	
	for _, id := range ids {
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notes/%s", baseURL, id), nil)
		if err != nil {
			continue
		}
		
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}
	
	fmt.Println("✓ Cleanup completed")
}

func TestNoteSpecificEndpoints(baseURL string, noteIDs []string) {
	if len(noteIDs) == 0 {
		fmt.Println("No note IDs available for specific endpoint testing")
		return
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	fmt.Println("\nTesting note-specific endpoints...")
	
	// Test GET /notes/{id}
	for _, id := range noteIDs[:min(3, len(noteIDs))] {
		resp, err := client.Get(fmt.Sprintf("%s/notes/%s", baseURL, id))
		if err != nil {
			fmt.Printf("  ✗ Failed to get note %s: %v\n", id, err)
		} else if resp.StatusCode == 200 {
			fmt.Printf("  ✓ Successfully retrieved note %s\n", id)
		} else {
			fmt.Printf("  ✗ Failed to get note %s (status: %d)\n", id, resp.StatusCode)
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CheckServiceHealth(baseURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Try multiple times
	for i := 0; i < 3; i++ {
		resp, err := client.Get(baseURL + "/healthz")
		if err == nil && resp.StatusCode == 204 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func main() {
	baseURL := "http://localhost:8080"
	
	fmt.Println("=========================================")
	fmt.Println("Load Test for Notes Service")
	fmt.Println("=========================================")
	
	// Check service health
	fmt.Printf("\nChecking service at %s...\n", baseURL)
	if !CheckServiceHealth(baseURL) {
		fmt.Printf("❌ Service is not healthy at %s\n", baseURL)
		fmt.Println("\nPlease make sure your service is running:")
		fmt.Println("  docker compose up -d")
		fmt.Println("  or")
		fmt.Println("  go run main.go")
		return
	}
	fmt.Println("✓ Service is healthy")
	
	// Create test notes
	fmt.Println("\n--- Setup Phase ---")
	testNoteIDs := CreateTestNotes(baseURL, 3)
	
	// Test note-specific endpoints
	if len(testNoteIDs) > 0 {
		TestNoteSpecificEndpoints(baseURL, testNoteIDs)
	}
	
	// Run load test
	fmt.Println("\n--- Load Test Phase ---")
	config := TestConfig{
		BaseURL:     baseURL,
		NumWorkers:  3,
		NumRequests: 30,
	}
	
	result := RunSimpleLoadTest(config)
	
	// Print results
	fmt.Println("\n--- Results ---")
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful: %d (%.1f%%)\n", result.Successful, 
		float64(result.Successful)/float64(result.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.1f%%)\n", result.Failed, 
		float64(result.Failed)/float64(result.TotalRequests)*100)
	
	fmt.Println("\nDetailed Endpoint Statistics:")
	for endpoint, count := range result.EndpointStats {
		fmt.Printf("  %s: %d requests\n", endpoint, count)
	}
	
	// Cleanup
	if len(testNoteIDs) > 0 {
		fmt.Println("\n--- Cleanup Phase ---")
		DeleteTestNotes(baseURL, testNoteIDs)
	}
	
	fmt.Println("\n✓ Load test completed successfully!")
}