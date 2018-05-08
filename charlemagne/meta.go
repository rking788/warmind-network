package charlemagne

// MetaResponse contains the fields returned from the Charlemagne API for meta endpoints
type MetaResponse struct {
	*BaseResponse
	*InnerMetaResponse `json:"response"`
}

// InnerMetaResponse is a wrapper around the inner response JSON tag. This is where
// the majority of the data is stored.
type InnerMetaResponse struct {
	WeekNumber        int            `json:"weekNumber"`
	ActivityHash      string         `json:"activityHash"`
	MembershipTypes   string         `json:"membershipTypes"` // Comma separated
	ModeTypes         string         `json:"modeTypes"`       // Comma separated
	TopKineticWeapons []*WeaponStats `json:"topKineticWeapons"`
	TopEnergyWeapons  []*WeaponStats `json:"topEnergyWeapons"`
	TopPowerWeapons   []*WeaponStats `json:"topPowerWeapons"`
}

// WeaponStats summarizes statistics about a particular weapon in the meta responses
type WeaponStats struct {
	WeaponID   int    `json:"weaponId"`
	WeaponName string `json:"weaponName"`
	TotalKills int    `json:"totalKills"`
	WeaponType string `json:"weaponType"`
}
