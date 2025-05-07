package faction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveContributors(
	ctx context.Context,
	userEnergy []UserGymEnergy,
) error {
	const query = `
	INSERT INTO user_gym_energy_log
		(torn_id, strength, speed, defense, dexterity, total, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (torn_id, timestamp) DO NOTHING
`

	for _, s := range userEnergy {
		_, err := r.db.Exec(
			ctx,
			query,
			s.UserID,
			s.Strength,
			s.Speed,
			s.Defense,
			s.Dexterity,
			s.Total,
			s.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("insert failed for user %s : %w", s.UserID, err)
		}
	}

	return nil
}
