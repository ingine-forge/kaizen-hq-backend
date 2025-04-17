package energy

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetLastEnergyRecord(ctx context.Context, tornID int64) (*time.Time, error) {
	var lastDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT date FROM energy_usage 
         WHERE torn_id = $1 
         ORDER BY date DESC LIMIT 1`,
		tornID).Scan(&lastDate)

	if err != nil {
		return nil, err
	}
	return &lastDate, nil
}

func (r *Repository) StoreDailyUsage(ctx context.Context, uuid uuid.UUID, tornID int64, energy float64, date time.Time) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO energy_usage (torn_id, date, energy, username)
        VALUES ($1, $2, $3,
            (SELECT username FROM users WHERE torn_id = $1)
        )
        ON CONFLICT (torn_id, date) DO UPDATE 
        SET energy = EXCLUDED.energy`,
		tornID, date.UTC().Truncate(24*time.Hour), energy)
	return err
}

// Add to Repository struct
func (r *Repository) GetUsersWithAPIKeys(ctx context.Context) ([]UserEnergyRecord, error) {
	const query = `SELECT torn_id, username, api_key FROM users WHERE api_key != ''`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserEnergyRecord
	for rows.Next() {
		var u UserEnergyRecord
		if err := rows.Scan(&u.TornID, &u.Username, &u.APIKey); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
