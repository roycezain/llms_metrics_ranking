package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/roycezain/llms_ranking/models"

	_ "github.com/lib/pq"
)

var db *sql.DB

type LLMRank struct {
	LLMName  string  `json:"llm_name"`
	AvgValue float64 `json:"avg_value"`
}

// LLMs List
var llms = []string{
	"GPT-4o", "Llama 3.1 405", "Mistral Large2", "Claude 3.5 Sonnet",
	"Gemini 1.5 Pro", "GPT-4o mini", "Llama 3.1 70B", "amba 1.5Large",
	"Mixtral 8x22B", "Gemini 1.5Flash", "Claude 3 Haiku", "Llama 3.1 8B",
}

// Metrics List
var metrics = []string{
	"TTFT", "TPS", "e2e_latency", "RPS",
}

// Initialize PostgreSQL client
func InitDB(connStr string) {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Set max number of open connections
	db.SetMaxOpenConns(25)

	// Set max idle connections
	db.SetMaxIdleConns(25)

	// Set max connection lifetime
	db.SetConnMaxLifetime(5 * time.Minute)

	//defer db.Close()

	// Create tables if not exist
	_, err = db.Exec(`

	    CREATE TABLE IF NOT EXISTS llms (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) NOT NULL
        );

        CREATE TABLE IF NOT EXISTS metrics (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) NOT NULL
        );

        CREATE TABLE IF NOT EXISTS llm_metrics (
            id SERIAL PRIMARY KEY,
            llm_name VARCHAR(255),
            metric_type VARCHAR(50),
            value FLOAT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)

	if err != nil {
		panic(err)
	}

	// Insert LLMs and metrics into database
	for _, llm := range llms {
		_, err = db.Exec("INSERT INTO llms (name) VALUES ($1)", llm)
		if err != nil {
			panic(err)
		}
	}

	for _, metric := range metrics {
		_, err = db.Exec("INSERT INTO metrics (name) VALUES ($1)", metric)
		if err != nil {
			panic(err)
		}
	}

	log.Println("Initial LLMs and Metrics seeded into database")

}

// FetchLLMs fetches the list of LLMs from the database
func FetchLLMs() ([]string, error) {
	rows, err := db.Query("SELECT name FROM llms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var llms []string
	for rows.Next() {
		var llm string
		if err := rows.Scan(&llm); err != nil {
			return nil, err
		}
		llms = append(llms, llm)
	}
	return llms, nil
}

// FetchMetrics fetches the list of LLMs from the database
func FetchMetrics() ([]string, error) {
	rows, err := db.Query("SELECT name FROM metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []string
	for rows.Next() {
		var metric string
		if err := rows.Scan(&metric); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// Batch Store Metrics inserts multiple metrics into the database in a single query.
func BatchStoreMetrics(metrics []models.LLMMetric, batchSize int) error {
	if len(metrics) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, batchSize)
	valueArgs := make([]interface{}, 0, batchSize*4)
	for i, metric := range metrics {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, metric.LLMName, metric.MetricType, metric.Value, metric.CreatedAt)
	}

	query := fmt.Sprintf("INSERT INTO llm_metrics (llm_name, metric_type, value, created_at) VALUES %s", strings.Join(valueStrings, ","))
	_, err := db.Exec(query, valueArgs...)
	return err
}

// Store llms Metrics
func StoreMetrics(metrics []models.LLMMetric) error {
	batchSize := 100 // Define batch size
	for i := 0; i < len(metrics); i += batchSize {
		end := i + batchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		batch := metrics[i:end]
		if err := BatchStoreMetrics(batch, batchSize); err != nil {
			return err
		}
	}
	return nil
}

// GetLLMRanking fetches and ranks LLMs by their mean metric values
func GetLLMRanking(metricType string) ([]LLMRank, error) {
	rows, err := db.Query(
		"SELECT llm_name, AVG(value) as avg_value FROM llm_metrics WHERE metric_type = $1 GROUP BY llm_name ORDER BY avg_value DESC",
		metricType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rankings []LLMRank
	for rows.Next() {
		var rank LLMRank
		if err := rows.Scan(&rank.LLMName, &rank.AvgValue); err != nil {
			return nil, err
		}
		rankings = append(rankings, rank)
	}
	return rankings, nil
}
