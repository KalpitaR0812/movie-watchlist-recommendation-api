package repositories

import "time"

// getCurrentTime returns the current UTC time
// This is a centralized helper function to avoid duplicate definitions
func getCurrentTime() time.Time {
	return time.Now().UTC()
}
