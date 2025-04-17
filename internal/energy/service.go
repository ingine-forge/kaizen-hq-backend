package energy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (s *Service) getEnergyUsage(apiKey string, from, to int64) (float64, error) {
	url := fmt.Sprintf("%suser/?key=%s&cat=125&from=%d&to=%d&selections=log",
		s.baseURL, apiKey, from, to)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	var total float64
	for _, entry := range data.Log {
		total += entry.Data.EnergyUsed
	}
	return total, nil
}

func (s *Service) getSignupTimestamp(apiKey string) (int64, error) {
	url := fmt.Sprintf("%suser/?key=%s&selections=profile", s.baseURL, apiKey)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var profile UserProfile

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return 0, err
	}

	// Parse Torn's datetime format
	signupTime, err := time.Parse("2006-01-02 15:04:05", profile.Signup)
	if err != nil {
		return 0, err
	}
	return signupTime.Unix(), nil
}

func (s *Service) ProcessUserEnergy(apiKey string, tornID int64) error {
	ctx := context.Background()

	// 1. Get user's last recorded date
	lastRecord, err := s.repo.GetLastEnergyRecord(ctx, tornID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to get last record: %w", err)
	}

	fmt.Println("last record found: ", lastRecord)

	// 2. Determine time range
	now := time.Now().UTC()
	yesterdayEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(-1 * time.Second)
	var fromTime time.Time

	if lastRecord == nil {
		// New user - fetch from signup
		signupTimestamp, err := s.getSignupTimestamp(apiKey)
		if err != nil {
			return fmt.Errorf("failed to get signup time: %w", err)
		}
		fromTime = time.Unix(signupTimestamp, 0)
	} else {
		// Existing user - fetch from day after last record
		fromTime = lastRecord.Add(24 * time.Hour)
	}

	// 3. Process each day separately
	for current := fromTime; current.Before(yesterdayEnd); current = current.Add(24 * time.Hour) {
		dayStart := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
		dayEnd := dayStart.Add(24*time.Hour - 1*time.Second)

		if dayEnd.After(yesterdayEnd) {
			dayEnd = yesterdayEnd
		}

		energy, err := s.getEnergyUsage(apiKey, dayStart.Unix(), dayEnd.Unix())
		if err != nil {
			return fmt.Errorf("failed to get energy for %s: %w", dayStart.Format("2006-01-02"), err)
		}

		uuidName := strconv.FormatInt(tornID, 10) + "_" + dayStart.Format("2006-01-02")
		uuidNameSpace := uuid.NameSpaceDNS
		deterministicUUID := uuid.NewSHA1(uuidNameSpace, []byte(uuidName))

		fmt.Println(deterministicUUID)
		if err := s.repo.StoreDailyUsage(ctx, deterministicUUID, tornID, energy, dayStart); err != nil {
			return fmt.Errorf("failed to store energy for %s: %w", dayStart.Format("2006-01-02"), err)
		}

		log.Printf("Processed %s: %.2f energy", dayStart.Format("2006-01-02"), energy)
	}

	return nil
}

func (s *Service) ProcessAllUsers() error {
	ctx := context.Background()

	users, err := s.repo.GetUsersWithAPIKeys(ctx)
	if err != nil {
		return err
	}

	var successCount int
	for _, user := range users {
		if err := s.ProcessUserEnergy(user.APIKey, user.TornID); err != nil {
			log.Printf("[%d] Failed to process: %v", user.TornID, err)
			continue
		}
		successCount++
	}

	log.Printf("Processed %d/%d users", successCount, len(users))
	return nil
}
