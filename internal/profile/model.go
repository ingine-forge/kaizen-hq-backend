package profile

import "time"

type Profile struct {
	TornID       int       `json:"torn_id"`
	Name         string    `json:"name"`
	Rank         string    `json:"rank"`
	Property     string    `json:"property"`
	Donator      int       `json:"donator"`
	ProfileImage string    `json:"profile_image"`
	Signup       time.Time `json:"signup"`
	Awards       int       `json:"awards"`
	Level        int       `json:"level"`
	Friends      int       `json:"friends"`
	Enemies      int       `json:"enemies"`
	Discord      string    `json:"discord"`
}
