package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/roycezain/llms_ranking/api"
	"github.com/roycezain/llms_ranking/cache"
	"github.com/roycezain/llms_ranking/config"
	"github.com/roycezain/llms_ranking/monitor"
	//"github.com/roycezain/llms_ranking/randomizer"
	//"github.com/roycezain/llms_ranking/storage"
)

func main() {

	// Load the configuration
	config.LoadConfig()

	//register prometheus metrics
	monitor.Init()

	// Initialize DB connection
	//storage.InitDB(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.AppConfig.DBHost, config.AppConfig.DBPort, config.AppConfig.DBUser, config.AppConfig.DBPassword, config.AppConfig.DBName))

	cache.InitRedis(config.AppConfig.REDISAddr, config.AppConfig.REDISPwd, 0)

	//randomizer.ParallelSimulation()

	mux := http.NewServeMux()

	//metrics collector for prometheus
	mux.HandleFunc("GET /metrics", promhttp.Handler().ServeHTTP)

	// Define Auth route
	mux.HandleFunc("GET /ranking/{metric}", api.GetLLMRanking)

	server := http.Server{
		Addr:    config.AppConfig.SERPort,
		Handler: api.ValidateAPIKey(mux),
	}

	// Start server
	log.Println("Server started on ", config.AppConfig.SERPort)
	log.Fatal(server.ListenAndServe())
}
