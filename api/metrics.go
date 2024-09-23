package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/roycezain/llms_ranking/cache"
	"github.com/roycezain/llms_ranking/config"
	"github.com/roycezain/llms_ranking/monitor"
)

func GetLLMRanking(w http.ResponseWriter, r *http.Request) {
	//metric := r.URL.Query().Get("metric")
	metric := r.PathValue("metric")
	if metric == "" {
		http.Error(w, "Metric query param is required", http.StatusBadRequest)
		monitor.CacheMisses.WithLabelValues(metric).Inc()
		return
	}

	start := time.Now()
	rankings, err := cache.GetLLMRankings(metric)
	if err != nil {
		http.Error(w, "Error fetching rankings", http.StatusInternalServerError)
		monitor.CacheMisses.WithLabelValues(metric).Inc()
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rankings)

	monitor.CacheHits.WithLabelValues(metric).Inc()

	monitor.ResponseTimeHistogram.WithLabelValues(metric).Observe(time.Since(start).Seconds())

	duration := time.Since(start)
	if duration.Seconds() > 2 {
		log.Printf("Warning: API response took %v seconds", duration.Seconds())
	}
}

func ValidateAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		// Handle other requests (espercially prometheus scraping)...
		switch r.URL.Path {
		case "/metrics":
			next.ServeHTTP(w, r)
			return
		default:
			if apiKey != config.AppConfig.APIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
