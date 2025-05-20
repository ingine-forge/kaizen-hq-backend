package faction

import (
	"context"
	"kaizen-hq/config"
	"kaizen-hq/internal/client"
	"time"
)

type Service struct {
	repo       *Repository
	config     *config.Config
	tornClient client.Client
}

func NewService(repo *Repository, cfg *config.Config, tornClient client.Client) *Service {
	return &Service{repo: repo, config: cfg, tornClient: tornClient}
}

func (s *Service) MergeAndSaveGymEnergy(
	strengthData, speedData, defenseData, dexterityData client.StatMap,
) error {
	userStats := map[string]*UserGymEnergy{}
	now := time.Now()

	merge := func(data client.StatMap, field string) {
		for _, users := range data {
			for userID, info := range users {

				if _, exists := userStats[userID]; !exists {
					userStats[userID] = &UserGymEnergy{UserID: userID, Timestamp: now}
				}
				switch field {
				case "strength":
					userStats[userID].Strength = info.Contributed
				case "speed":
					userStats[userID].Speed = info.Contributed
				case "defense":
					userStats[userID].Defense = info.Contributed
				case "dexterity":
					userStats[userID].Dexterity = info.Contributed
				}
			}
		}
	}

	merge(strengthData, "strength")
	merge(speedData, "speed")
	merge(defenseData, "defense")
	merge(dexterityData, "dexterity")

	// Convert to slice for repo
	var energyList []UserGymEnergy
	for _, entry := range userStats {
		entry.Total = entry.Strength + entry.Speed + entry.Defense + entry.Dexterity
		energyList = append(energyList, *entry)
	}

	return s.repo.SaveContributors(context.Background(), energyList)
}

func (s *Service) UpdateGymEnergy(apiKey string) error {
	ctx := context.Background()

	strengthData, err := s.tornClient.FetchGymEnergy(ctx, apiKey, "gymstrength")
	if err != nil {
		return err
	}

	speedData, err := s.tornClient.FetchGymEnergy(ctx, apiKey, "gymspeed")
	if err != nil {
		return err
	}

	defenseData, err := s.tornClient.FetchGymEnergy(ctx, apiKey, "gymdefense")
	if err != nil {
		return err
	}

	dexterityData, err := s.tornClient.FetchGymEnergy(ctx, apiKey, "gymdexterity")
	if err != nil {
		return err
	}

	return s.MergeAndSaveGymEnergy(strengthData, speedData, defenseData, dexterityData)
}
