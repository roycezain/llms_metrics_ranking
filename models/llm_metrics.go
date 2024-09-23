package models

import (
	"time"
)

type LLMMetric struct {
	ID         int       `json:"id"`
	LLMName    string    `json:"llm_name"`
	MetricType string    `json:"metric_type"`
	Value      float64   `json:"value"`
	CreatedAt  time.Time `json:"created_at"`
}
