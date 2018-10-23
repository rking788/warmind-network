package models

// EquipmentBucket is the type of the key for the bucket type hash lookup
type EquipmentBucket uint

// Equipment bucket type definitions
const (
	_ EquipmentBucket = iota
	Kinetic
	Energy
	Power
	Ghost
	Helmet
	Gauntlets
	Chest
	Legs
	ClassArmor
	Artifact
	Subclass
)

func (bucket EquipmentBucket) String() string {
	switch bucket {
	case Kinetic:
		return "Kinetic"
	case Energy:
		return "Energy"
	case Power:
		return "Power"
	case Ghost:
		return "Ghost"
	case Helmet:
		return "Helmet"
	case Gauntlets:
		return "Gauntlets"
	case Chest:
		return "Chest"
	case Legs:
		return "Legs"
	case ClassArmor:
		return "ClassArmor"
	case Artifact:
		return "Artifact"
	case Subclass:
		return "Subclass"
	}

	return ""
}

// Loadout will hold all currently equipped items indexed by the equipment bucket for
// which they are equipped.
type Loadout map[EquipmentBucket]*Item

func NewLoadout() Loadout {
	return make(map[EquipmentBucket]*Item)
}

// Equipment is very similar to the Loadout type but contains all items in an
// equipment bucket instead of just the currently equipped item
type Equipment map[EquipmentBucket][]*Item

func NewEquipment() Equipment {
	e := make(Equipment)
	for i := Kinetic; i <= Subclass; i++ {
		// Equipment buckets should really only have 10 items total, at least that is
		// how character inventories/equipment slots work at the time of this writing
		e[i] = make([]*Item, 0, 10)
	}

	return e
}

// PersistedItem represents the data from a specific Item entry that should be
// persisted. Not all of the Item data should be stored as most of it
// could/will change by the time it is read back from persistent storage
type PersistedItem struct {
	ItemHash   uint   `json:"itemHash"`
	InstanceID string `json:"itemInstanceId"`
	BucketHash uint   `json:"bucketHash"`
}

// PersistedLoadout stores all of the PersistedItem entries that should be saved
// for a particular loadout.
type PersistedLoadout map[EquipmentBucket]*PersistedItem

// CalculateLightLevel will find the power level for the specific loadout.
// In Forsaken, power level is just a simple average now instead of a
// weighted average in previous versions
func (l Loadout) CalculateLightLevel() float64 {

	light := float64(l[Kinetic].Power()+
		l[Energy].Power()+
		l[Power].Power()+
		l[Helmet].Power()+
		l[Gauntlets].Power()+
		l[Chest].Power()+
		l[Legs].Power()+
		l[ClassArmor].Power()) / 8.0

	return light
}

func (l Loadout) ToSlice() []*Item {

	result := make([]*Item, 0, ClassArmor-Kinetic)
	for i := Kinetic; i <= ClassArmor; i++ {
		item := l[i]
		if item != nil {
			result = append(result, l[i])
		}
	}

	return result
}

func (l Loadout) ToPersistedLoadout() PersistedLoadout {

	persisted := make(PersistedLoadout)
	for equipmentBucket, item := range l {
		persisted[equipmentBucket] = &PersistedItem{
			ItemHash:   item.ItemHash,
			InstanceID: item.InstanceID,
			BucketHash: item.BucketHash,
		}

	}

	return persisted
}
