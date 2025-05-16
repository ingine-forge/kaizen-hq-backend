package permission

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	query := `INSERT INTO permissions (name, description) VALUES ($1, $2) RETURNING id`

	err := r.db.QueryRow(ctx, query, permission.Name, permission.Description).Scan(&permission.ID)

	return permission, err
}

func (r *Repository) GetPermissionByName(ctx context.Context, name string) (*Permission, error) {
	permission := &Permission{}

	query := `SELECT * FROM permissions WHERE name = $1`
	err := r.db.QueryRow(ctx, query, name).Scan(permission)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("no permission with that name found")
		}
		return nil, err
	}

	return permission, nil
}
