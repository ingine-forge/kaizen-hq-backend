package energy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Service struct {
	repo       *Repository
	httpClient *http.Client
	baseURL    string
}

func NewService(repo *Repository, baseURL string) *Service {
	return &Service{
		repo:       repo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

func (s *Service) FetchUserEnergy(apiKey string) (float64, error) {
	url := s.baseURL + "user?key=" + apiKey + "&cat=125&selections=log&comment=KaizenHQ"

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return s.calculateDailyEnergy(data), nil
}

func (s *Service) calculateDailyEnergy(data LogResponse) float64 {
	var total float64
	// now := time.Now().UTC()
	// todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for _, entry := range data.Log {
		// if time.Unix(entry.Timestamp, 0).After(todayStart) {
		total += entry.Data.EnergyUsed
		// }
	}

	fmt.Println(total)

	return total
}

func (s *Service) ProcessAllUsers() error {
	ctx := context.Background()

	// 1. Get all users with API keys
	users, err := s.repo.GetUsersWithAPIKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	// 2. Process each user
	var successCount int
	for _, user := range users {
		energy, err := s.FetchUserEnergy(user.APIKey)
		if err != nil {
			log.Printf("[%s] Failed to fetch energy: %v", user.Username, err)
			continue
		}

		// 3. Store the result
		if err := s.repo.StoreDailyUsage(ctx, user.TornID, energy, time.Now().UTC()); err != nil {
			log.Printf("[%s] Failed to store energy: %v", user.Username, err)
			continue
		}

		successCount++
		log.Printf("[%s] Tracked %.2f energy", user.Username, energy)
	}

	log.Printf("Energy tracking complete: %d/%d users processed", successCount, len(users))
	return nil
}
