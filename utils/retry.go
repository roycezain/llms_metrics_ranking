package utils

import (
	"errors"
	"log"
	"time"
)

func RetryOperation(operation func() error, retries int, delay time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		err = operation()
		if err == nil {
			return nil // Success, exit the loop
		}
		log.Printf("Attempt %d/%d failed: %v. Retrying in %v...", i+1, retries, err, delay)
		time.Sleep(delay) // Wait before retrying
	}
	return errors.New("operation failed after max retries: " + err.Error())

}
