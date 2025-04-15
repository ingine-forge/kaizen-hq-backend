package tornapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles API calls to the torn service
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new API Client
func NewClient() *Client {
	return &Client{
		BaseURL: "https://api.torn.com",
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// BaseStat represents common properties for all stat types
type BaseStat struct {
	Before    string  `json:"before"`
	After     float64 `json:"after"`
	Increased float64 `json:"increased"`
}

// BaseGymData contains common fields for all gym training types
type BaseGymData struct {
	Trains     int `json:"trains"`
	EnergyUsed int `json:"energy_used"`
	HappyUsed  int `json:"happy_used"`
	Gym        int `json:"gym"`
}

// GymStatType identifies which stat is being trained
type GymStatType string

const (
	StatTypeStrength  GymStatType = "strength"
	StatTypeSpeed     GymStatType = "speed"
	StatTypeDefense   GymStatType = "defense"
	StatTypeDexterity GymStatType = "dexterity"
)

// Log represents the entire log object from the api
type Log struct {
	Log LogEntry `json:"log"`
}

type GymLog struct {
	ID           string          `json:"-"`
	LogID        int             `json:"-"`
	Title        string          `json:"title"`
	Timestamp    time.Time       `json:"-"`
	Category     string          `json:"category"`
	RawData      json.RawMessage `json:"-"`
	StatType     GymStatType     `json:"-"`
	BaseData     BaseGymData     `json:"-"`
	Stats        BaseStat        `json:"-"`
	Params       LogEntryParams  `json:"params"`
	RawTimestamp int64           `json:"timestamp"`
}

// LogEntry represents a single activity log inside Log
type LogEntry struct {
	Log       int             `json:"log"`
	Title     string          `json:"title"`
	Timestamp int64           `json:"timestamp"`
	Category  string          `json:"category"`
	Data      json.RawMessage `json:"data"`
	Params    LogEntryParams  `json:"params"`
}

// LogEntryData represents the data inside the LogEntry
// This only uses the energy_used in the gym
// TODO: Have multiple types of LogEntryData
// type LogEntryData struct {
// 	Trains     int `json:"trains"`
// 	EnergyUsed int `json:"energy_used"`
// }

// LogEntryParams represents the params inside the LogEntry
type LogEntryParams struct {
	Color string `json:"color"`
}

type LogResponse struct {
	Log map[string]LogEntry `json:"log"`
}

// GetGymLogs parses the JSON response and returns only gym logs
func (c *Client) GetGymLogs(apiKey string, fromTime time.Time, toTime time.Time) ([]GymLog, error) {
	// build the API URL with parameters
	url := fmt.Sprintf("%s/user?selections=log&key=%s&cat=125", c.BaseURL, apiKey)

	// Make the request
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// check for error status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var response LogResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal log response: %w", err)
	}

	// Extract gym logs
	gymLogs := []GymLog{}

	for logID, entry := range response.Log {
		// Filter only for "Gym" category logs
		if entry.Category == "Gym" {
			// Apply time filters if needed
			logTime := time.Unix(entry.Timestamp, 0)
			if (!fromTime.IsZero() && logTime.Before(fromTime)) || (!toTime.IsZero() && logTime.After(toTime)) {
				continue
			}

			// Create a gym log with basic info
			gymLog := GymLog{
				ID:           logID,
				LogID:        entry.Log,
				Title:        entry.Title,
				Timestamp:    logTime,
				Category:     entry.Category,
				RawData:      entry.Data,
				Params:       entry.Params,
				RawTimestamp: entry.Timestamp,
			}

			// First unmarshal common fields
			var baseData BaseGymData
			if err := json.Unmarshal(entry.Data, &baseData); err != nil {
				continue // Skip this entry if we can't parse the base data
			}
			gymLog.BaseData = baseData

			// Determine stat type and extract stat data
			gymLog.StatType = determineStatType(entry.Data)
			gymLog.Stats = extractStats(entry.Data, gymLog.StatType)

			gymLogs = append(gymLogs, gymLog)

		}
	}
	return gymLogs, nil

}

// determineStatType detects which stat type is present in the JSON data
func determineStatType(data json.RawMessage) GymStatType {
	// Try each possible stat type
	statTypes := []GymStatType{
		StatTypeStrength,
		StatTypeSpeed,
		StatTypeDefense,
		StatTypeDexterity,
	}

	for _, statType := range statTypes {
		var checkMap map[string]interface{}
		if err := json.Unmarshal(data, &checkMap); err != nil {
			continue
		}

		beforeKey := string(statType) + "_before"
		if _, exists := checkMap[beforeKey]; exists {
			return statType
		}
	}

	return ""
}

// extractStats reads the stat values for the determined type
func extractStats(data json.RawMessage, statType GymStatType) BaseStat {
	var result BaseStat
	if statType == "" {
		return result
	}

	var rawData map[string]any
	if err := json.Unmarshal(data, &rawData); err != nil {
		return result
	}

	// Extract using the appropriate field names
	beforeKey := string(statType) + "_before"
	afterKey := string(statType) + "_after"
	increasedKey := string(statType) + "_increased"

	if beforeVal, ok := rawData[beforeKey]; ok {
		if beforeStr, ok := beforeVal.(string); ok {
			result.Before = beforeStr
		}
	}

	if afterVal, ok := rawData[afterKey]; ok {
		if afterNum, ok := afterVal.(float64); ok {
			result.After = afterNum
		}
	}

	if increasedVal, ok := rawData[increasedKey]; ok {
		if increasedNum, ok := increasedVal.(float64); ok {
			result.Increased = increasedNum
		}
	}

	return result
}
