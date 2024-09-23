package randomizer

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/roycezain/llms_ranking/cache"
	"github.com/roycezain/llms_ranking/config"
	"github.com/roycezain/llms_ranking/models"
	"github.com/roycezain/llms_ranking/storage"
	"github.com/roycezain/llms_ranking/utils"
)

// LLMs List
/*var llms = []string{
	"GPT-4o", "Llama 3.1 405", "Mistral Large2", "Claude 3.5 Sonnet",
	"Gemini 1.5 Pro", "GPT-4o mini", "Llama 3.1 70B", "amba 1.5Large",
	"Mixtral 8x22B", "Gemini 1.5Flash", "Claude 3 Haiku", "Llama 3.1 8B",
} */

// Metrics List
/*var metrics = []string{
	"TTFT", "TPS", "e2e_latency", "RPS",
}*/

var numDataPoints int = 1000

// SimulateLLM simulates a single LLM and generates metrics.
func SimulateLLM(llmName string, metricType string, count int) []models.LLMMetric {
	metrics := make([]models.LLMMetric, count)
	for i := 0; i < count; i++ {
		metrics[i] = models.LLMMetric{
			LLMName:    llmName,
			MetricType: metricType,
			Value:      generateRandomValue(metricType),
			CreatedAt:  time.Now(),
		}
	}
	return metrics
}

// ParallelSimulation runs simulations for multiple LLMs concurrently
func ParallelSimulation() error {
	var wg sync.WaitGroup
	llmChannel := make(chan []models.LLMMetric)

	// Worker to batch insert metrics into the database
	go func() {
		for batch := range llmChannel {
			if err := RetryableStoreMetric(batch, config.AppConfig.MaxRetries, time.Duration(config.AppConfig.RetryDelayMs)); err != nil {
				// Handle error: retry logic or log failure
				log.Println("Error storing batch: ", err)
			}
		}
	}()

	var (
		metricsy []string
		llmsy    []string
	)
	//var metricsy []string
	metricsy, err := cache.GetMetrics()
	if err != nil {
		return err
	}

	//var llmsy []string
	llmsy, err = cache.GetLLMS()
	if err != nil {
		return err
	}

	for _, llm := range llmsy {
		for _, metric := range metricsy {
			wg.Add(1)
			go func(llm, metric string) {
				defer wg.Done()
				simulatedMetrics := SimulateLLM(llm, metric, numDataPoints)
				llmChannel <- simulatedMetrics
			}(llm, metric)
		}
	}

	// Wait for all simulations to complete
	wg.Wait()

	// Close the channel when all simulations are done
	close(llmChannel)
	return nil
}

// RetryableStoreMetric tries to store the metric in the database with retry logic.
func RetryableStoreMetric(batch []models.LLMMetric, maxRetries int, delay time.Duration) error {
	operation := func() error {
		// Attempt to store the metric
		storage.StoreMetrics(batch)
		return nil
	}

	// Use RetryOperation with maxRetries and delay
	err := utils.RetryOperation(operation, maxRetries, delay)
	if err != nil {
		log.Printf("Failed to store metric after retries: %v", err)
		return err
	}
	return nil
}

func generateRandomValue(metricType string) float64 {
	rand.New(rand.NewSource(config.AppConfig.Seed)) // Seed globally set once for efficiency
	switch metricType {
	case "TTFT":
		return rand.Float64() * 100 // Simulate Time to First Token in ms
	case "TPS":
		return rand.Float64() * 1000 // Simulate Tokens Per Second
	case "e2e_latency":
		return rand.Float64() * 500 // Simulate end-to-end latency in ms
	case "RPS":
		return rand.Float64() * 50 // Simulate Requests Per Second
	default:
		return rand.Float64() * 1000
	}
}

func AfterNewDataSimulation(metric string) error {
	// Invalidate the cache after new simulations
	err := cache.InvalidateLLMRankingsCache(metric)
	if err != nil {
		log.Println("Error invalidating cache:", err)
	}

	// Optionally, re-populate the cache for immediate access
	rankings, err := storage.GetLLMRanking(metric)
	if err != nil {
		log.Println("Error fetching rankings from DB:", err)
	}
	err = cache.CacheRankings("llm_rankings:"+metric, rankings)
	if err != nil {
		log.Println("Error caching rankings:", err)
	}

	return nil
}
