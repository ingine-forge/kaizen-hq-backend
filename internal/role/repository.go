package role

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

func (r *Repository) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	query := `INSERT INTO roles (name, description, is_leadership) VALUES ($1, $2, $3) RETURNING id`

	err := r.db.QueryRow(ctx, query, role.Name, role.Description, role.IsLeadership).Scan(&role.ID)

	return role, err
}

func (r *Repository) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	role := &Role{}

	query := `SELECT * FROM roles WHERE name = $1`
	err := r.db.QueryRow(ctx, query, name).Scan(role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("no role with that name found")
		}
		return nil, err
	}

	return role, nil
}

func (r *Repository) AssignPermission(ctx context.Context, roleID int, permissionID int) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`

	_, err := r.db.Exec(ctx, query, roleID, permissionID)
	return err
}
