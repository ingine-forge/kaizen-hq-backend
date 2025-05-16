package torn

import "time"

type StatMap map[string]map[string]ContributorInfo

type ContributorInfo struct {
	Contributed int `json:"contributed"`
	InFaction   int `json:"in_faction"`
}

type Profile struct {
	Name         string         `json:"name"`
	Rank         string         `json:"rank"`
	Property     string         `json:"property"`
	Donator      int            `json:"donator"`
	PlayerID     int            `json:"player_id"`
	ProfileImage string         `json:"profile_image"`
	Faction      ProfileFaction `json:"faction"`
	Signup       time.Time      `json:"signup"`
	Awards       int            `json:"awards"`
	Level        int            `json:"level"`
	Friends      int            `json:"friends"`
	Enemies      int            `json:"enemies"`
}

type ProfileFaction struct {
	FactionName   string `json:"faction_name"`
	DaysInFaction int    `json:"days_in_faction"`
	Position      string `json:"position"`
}

type Discord struct {
	UserID    int    `json:"userID"`
	DiscordID string `json:"discordID"`
}
