package models

import "time"

// Profile contains all information about a specific Destiny membership, including character and
// inventory information.
type Profile struct {
	MembershipType        int
	MembershipID          string
	BungieNetMembershipID string
	DateLastPlayed        time.Time
	DisplayName           string
	Characters            CharacterList

	AllItems []*Item

	// A map between the character ID and the Loadout currently equipped on that character
	Loadouts map[string]Loadout

	Equipments map[string]Equipment
	// A list of the current and available character activities sorted by most
	// recent activity start date
	Activities []*CharacterActivities

	// NOTE: Still not sure this is the best approach to flatten items into a single list,
	// it works well for now so we will go with it. There are too many potential spots to
	// look for an item.
	//Equipments       map[string]ItemList
	//Inventories      map[string]ItemList
	//ProfileInventory ItemList
	//Currencies       ItemList
}
