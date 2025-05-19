package torn

type StatMap map[string]map[string]ContributorInfo

type ContributorInfo struct {
	Contributed int `json:"contributed"`
	InFaction   int `json:"in_faction"`
}

type User struct {
	Rank         string `json:"rank"`
	Level        int    `json:"level"`
	Honor        int    `json:"honor"`
	Gender       string `json:"gender"`
	Property     string `json:"property"`
	Signup       string `json:"signup"`
	Awards       int    `json:"awards"`
	Friends      int    `json:"friends"`
	Enemies      int    `json:"enemies"`
	ForumPosts   int    `json:"forum_posts"`
	Karma        int    `json:"karma"`
	Age          int    `json:"age"`
	Role         string `json:"role"`
	Donator      int    `json:"donator"`
	PlayerID     int    `json:"player_id"`
	Name         string `json:"name"`
	PropertyID   int    `json:"property_id"`
	Revivable    int    `json:"revivable"`
	ProfileImage string `json:"profile_image"`
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
