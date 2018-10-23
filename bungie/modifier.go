package bungie

import "github.com/rking788/warmind-network/models"

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
	GetBuffStatus(item *models.Item) BuffStatus
}

// Modifier represents a specific modifier that can be applied to an activity. Example: Void Singe
// This struct should provide a callback to indicate the BuffStatus that this modifier
// would cause for a specific item.
type Modifier struct {
	Hash        uint
	Name        string
	Description string
	BuffStatus  func(*models.Item) BuffStatus
}

var (
	modifierLookup = map[uint]Modifier{
		solarSingeHash: Modifier{
			Hash:        solarSingeHash,
			Name:        "Solar Singe",
			Description: "",
			BuffStatus:  SolarSinge,
		},
		arcSingeHash: Modifier{
			Hash:        arcSingeHash,
			Name:        "Arc Singe",
			Description: "",
			BuffStatus:  ArcSinge,
		},
		voidSingeHash: Modifier{
			Hash:        voidSingeHash,
			Name:        "Void Singe",
			Description: "",
			BuffStatus:  VoidSinge,
		},
	}
)

// SolarSinge will provide a buff to items that have a solar damage type
// and no buff to any other items.
func SolarSinge(i *models.Item) BuffStatus {
	if i.Damage() == solarDamage {
		return Buff
	}

	return NoBuff
}

// ArcSinge will provide a buff to items that have a damage type of Arc and no buff to others
func ArcSinge(i *models.Item) BuffStatus {
	if i.Damage() == arcDamage {
		return Buff
	}

	return NoBuff
}

// VoidSinge will provide a buff to items with a void damage type and no buff to others
func VoidSinge(i *models.Item) BuffStatus {
	if i.Damage() == voidDamage {
		return Buff
	}

	return NoBuff
}
