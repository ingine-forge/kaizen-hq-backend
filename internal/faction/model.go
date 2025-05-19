package faction

import (
	"time"
)

type UserGymEnergy struct {
	UserID    string
	Strength  int
	Speed     int
	Defense   int
	Dexterity int
	Total     int
	Timestamp time.Time
}

type Faction struct {
	ID        int    `json:"ID"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	TagImage  string `json:"tag_image"`
	Leader    int    `json:"leader"`
	CoLeader  int    `json:"co-leader"`
	Respect   int    `json:"respect"`
	Age       int    `json:"age"`
	Capacity  int    `json:"capacity"`
	BestChain int    `json:"best_chain"`
}

type Position struct {
	Default                 int `json:"default"`
	CanUseMedicalItem       int `json:"canUseMedicalItem"`
	CanUseBoosterItem       int `json:"canUseBoosterItem"`
	CanUseDrugItem          int `json:"canUseDrugItem"`
	CanUseEnergyRefill      int `json:"canUseEnergyRefill"`
	CanUseNerveRefill       int `json:"canUseNerveRefill"`
	CanLoanTemporaryItem    int `json:"canLoanTemporaryItem"`
	CanLoanWeaponAndArmory  int `json:"canLoanWeaponAndArmory"`
	CanRetrieveLoanedArmory int `json:"canRetrieveLoanedArmory"`
	CanAccessFactionAPI     int `json:"canAccessFactionApi"`
	CanGiveItem             int `json:"canGiveItem"`
	CanGiveMoney            int `json:"canGiveMoney"`
	CanGivePoints           int `json:"canGivePoints"`
	CanManageForum          int `json:"canManageForum"`
	CanManageApplications   int `json:"canManageApplications"`
	CanKickMembers          int `json:"canKickMembers"`
	CanAdjustMemberBalance  int `json:"canAdjustMemberBalance"`
	CanManageWars           int `json:"canManageWars"`
	CanManageUpgrades       int `json:"canManageUpgrades"`
	CanSendNewsletter       int `json:"canSendNewsletter"`
	CanChangeAnnouncement   int `json:"canChangeAnnouncement"`
	CanChangeDescription    int `json:"canChangeDescription"`
	CanManageOC2            int `json:"canManageOC2"`
}
