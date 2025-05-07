package faction

import "time"

type UserGymEnergy struct {
	UserID    string
	Strength  int
	Speed     int
	Defense   int
	Dexterity int
	Total     int
	Timestamp time.Time
}
