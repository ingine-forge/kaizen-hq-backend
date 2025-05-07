package torn

type StatMap map[string]map[string]ContributorInfo

type ContributorInfo struct {
	Contributed int `json:"contributed"`
	InFaction   int `json:"in_faction"`
}
