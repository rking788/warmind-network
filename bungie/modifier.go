package bungie

// BuffStatus indicates the postiive or negative effect an item/modifier/activity has on a
// player's character overall.
type BuffStatus int

const (
	// Buff indicats a positive enhacement to the player
	Buff BuffStatus = 1
	// NoBuff is the default and means no effect
	NoBuff BuffStatus = 0
	// Debuff indicates a negative effect on the player (decreased ability recharge for example)
	Debuff BuffStatus = -1
)

// ItemBuffProvider is an interface that a modifier should implement to return the BuffStatus that
// a modifier has on an item. Buff for matching elements for example.
type ItemBuffProvider interface {
	GetBuffStatus(item *Item) BuffStatus
}

// Modifier represents a specific modifier that can be applied to an activity. Example: Void Singe
// This struct should provide a callback to indicate the BuffStatus that this modifier
// would cause for a specific item.
type Modifier struct {
	Hash        uint
	Name        string
	Description string
	BuffStatus  func(*Item) BuffStatus
}

var (
	modifierLookup = map[uint]Modifier{
		2558957669: Modifier{
			Hash:        2558957669,
			Name:        "Solar Singe",
			Description: "",
			BuffStatus:  SolarSinge,
		},
		3215384520: Modifier{
			Hash:        3215384520,
			Name:        "Arc Singe",
			Description: "",
			BuffStatus:  ArcSinge,
		},
		3362074814: Modifier{
			Hash:        3362074814,
			Name:        "Void Singe",
			Description: "",
			BuffStatus:  VoidSinge,
		},
	}
)

// SolarSinge will provide a buff to items that have a solar damage type
// and no buff to any other items.
func SolarSinge(i *Item) BuffStatus {
	if i.Damage() == SolarDamage {
		return Buff
	}

	return NoBuff
}

// ArcSinge will provide a buff to items that have a damage type of Arc and no buff to others
func ArcSinge(i *Item) BuffStatus {
	if i.Damage() == ArcDamage {
		return Buff
	}

	return NoBuff
}

// VoidSinge will provide a buff to items with a void damage type and no buff to others
func VoidSinge(i *Item) BuffStatus {
	if i.Damage() == VoidDamage {
		return Buff
	}

	return NoBuff
}
