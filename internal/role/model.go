package role

type Role struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsLeadership bool   `json:"is_leadership"`
}
