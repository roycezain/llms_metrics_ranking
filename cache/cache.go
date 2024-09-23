package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/roycezain/llms_ranking/storage"
)

const (
	redisRankingTTL     = 10 * time.Minute
	redisLLMSMetricsTTL = 30 * time.Minute
)

var (
	ctx = context.Background() // Global Redis context
	rdb *redis.Client
)

// Initialize Redis client
func InitRedis(addr, password string, db int) error {
	if addr == "" || password == "" {
		return errors.New("redis address or password is empty")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Ping Redis to ensure connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// Data structure to store ranking information
type LLMRanking struct {
	LLMName  string  `json:"llm_name"`
	AvgValue float64 `json:"avg_value"`
}

///////////////// GET METRICS FROM REDIS CACHE ////////

func GetMetrics() ([]string, error) {

	cacheKey := "metrics:"

	// Check Redis cache
	cachedMetrics, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { // Cache miss
		log.Println("Cache miss, querying PostgreSQL")

		var metrics []string
		metrics, err := storage.FetchMetrics()
		if err != nil {
			log.Fatalf("Error fetching LLMs from postgres: %v", err)
		}

		// Cache the result in Redis with a TTL of 10 minutes
		err = CacheMetrics(cacheKey, metrics)
		if err != nil {
			log.Println("Error caching data in Redis:", err)
		}

		return metrics, nil

	} else if err != nil {
		return nil, err
	}

	// Cache hit, unmarshal the result
	var metrics []string
	err = json.Unmarshal([]byte(cachedMetrics), &metrics)
	if err != nil {
		return nil, err
	}

	log.Println("Cache hit, returning data from Redis")
	return metrics, nil
}

// Cache the LLM rankings in Redis with a TTL
func CacheMetrics(cacheKey string, llms []string) error {
	metricsData, err := json.Marshal(llms)
	if err != nil {
		return err
	}

	// Set a TTL of 10 minutes for the cache
	err = rdb.Set(ctx, cacheKey, metricsData, redisLLMSMetricsTTL).Err()
	return err
}

///////////////// GET LLMS FROM REDIS CACHE ////////

func GetLLMS() ([]string, error) {

	cacheKey := "llms:"

	// Check Redis cache
	cachedLLMs, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { // Cache miss
		log.Println("Cache miss, querying PostgreSQL")

		var llms []string
		llms, err := storage.FetchLLMs()
		if err != nil {
			log.Fatalf("Error fetching LLMs from postgres: %v", err)
		}

		// Cache the result in Redis with a TTL of 10 minutes
		err = CacheLLMs(cacheKey, llms)
		if err != nil {
			log.Println("Error caching data in Redis:", err)
		}

		return llms, nil

	} else if err != nil {
		return nil, err
	}

	// Cache hit, unmarshal the result
	var llms []string
	err = json.Unmarshal([]byte(cachedLLMs), &llms)
	if err != nil {
		return nil, err
	}

	log.Println("Cache hit, returning data from Redis")
	return llms, nil
}

// Cache the LLM rankings in Redis with a TTL
func CacheLLMs(cacheKey string, llms []string) error {
	llmsData, err := json.Marshal(llms)
	if err != nil {
		return err
	}

	// Set a TTL of 10 minutes for the cache
	err = rdb.Set(ctx, cacheKey, llmsData, redisLLMSMetricsTTL).Err()
	return err
}

// Function to get rankings from Redis cache or PostgreSQL
func GetLLMRankings(metric string) ([]storage.LLMRank, error) {

	if metric == "" {
		return nil, errors.New("metric is empty")
	}
	cacheKey := "llm_rankings:" + metric

	// Check Redis cache
	cachedRankings, err := rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { // Cache miss
		log.Println("Cache miss, querying PostgreSQL")

		// Fetch rankings from PostgreSQL
		rankings, err := storage.GetLLMRanking(metric)
		if err != nil {
			return nil, err
		}

		// Cache the result in Redis with a TTL of 10 minutes
		err = CacheRankings(cacheKey, rankings)
		if err != nil {
			log.Println("Error caching data in Redis:", err)
		}

		return rankings, nil

	} else if err != nil {
		return nil, err
	}

	// Cache hit, unmarshal the result
	var rankings []storage.LLMRank
	err = json.Unmarshal([]byte(cachedRankings), &rankings)
	if err != nil {
		return nil, err
	}

	log.Println("Cache hit, returning data from Redis")
	return rankings, nil
}

// Cache the LLM rankings in Redis with a TTL
func CacheRankings(cacheKey string, rankings []storage.LLMRank) error {
	rankingData, err := json.Marshal(rankings)
	if err != nil {
		return err
	}

	// Set a TTL of 10 minutes for the cache
	err = rdb.Set(ctx, cacheKey, rankingData, redisRankingTTL).Err()
	return err
}

// Invalidate cache for a specific metric
func InvalidateLLMRankingsCache(metric string) error {
	cacheKey := "llm_rankings:" + metric
	err := rdb.Del(ctx, cacheKey).Err()
	return err
}
