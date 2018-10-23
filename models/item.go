package models

import (
	"fmt"
)

// DamageType will represent the specific damage type tied to a weapon or NoDamage if an item does
// not deal any damage.
type DamageType uint

//Item represents a single inventory item returned usually by the GetProfile endpoint
type Item struct {
	// DestinyItemComponent https://bungie-net.github.io/multi/schema_Destiny-Entities-Items-DestinyItemComponent.html#schema_Destiny-Entities-Items-DestinyItemComponent
	ItemHash       uint   `json:"itemHash"`
	InstanceID     string `json:"itemInstanceId"`
	BucketHash     uint   `json:"bucketHash"`
	Lockable       bool   `json:"lockable"`
	BindStatus     int    `json:"bindStatus"`
	State          int    `json:"state"`
	Location       int    `json:"location"`
	TransferStatus int    `json:"transferStatus"`
	Quantity       int    `json:"quantity"`
	*ItemInstance
	*Character
}

// ItemInstance will hold information about a specific instance of an instanced item, this can include item stats,
// perks, etc. as well as equipped status and things like that.
type ItemInstance struct {
	//https://bungie-net.github.io/multi/schema_Destiny-Entities-Items-DestinyItemInstanceComponent.html#schema_Destiny-Entities-Items-DestinyItemInstanceComponent
	IsEquipped         bool `json:"isEquipped"`
	CanEquip           bool `json:"canEquip"`
	Quality            int  `json:"quality"`
	CannotEquipReason  int  `json:"cannotEquipReason"`
	DamageType         `json:"damageTypeHash"`
	EquipRequiredLevel int `json:"equipRequiredLevel"`
	PrimaryStat        *struct {
		//https://bungie-net.github.io/multi/schema_Destiny-DestinyStat.html#schema_Destiny-DestinyStat
		StatHash     uint `json:"statHash"`
		Value        int  `json:"value"`
		MaximumValue int  `json:"maximumValue"`
		ItemLevel    int  `json:"itemLevel"`
	} `json:"primaryStat"`
}

// ItemMetadata is responsible for holding data from the manifest in-memory that is used often
// when interacting wth different character's inventories. These values are used so much
// that it would be a big waste of time to query the manifest data from the DB for every use.
type ItemMetadata struct {
	TierType   int
	ClassType  int
	BucketHash uint
}

func (i *Item) String() string {
	if i.ItemInstance != nil {
		if i.ItemInstance.PrimaryStat != nil {
			return fmt.Sprintf("Item{itemHash: %d, itemID: %s, light:%d, isEquipped: %v, quantity: %d}", i.ItemHash, i.InstanceID, i.PrimaryStat.Value, i.IsEquipped, i.Quantity)
		}

		return fmt.Sprintf("Item{itemHash: %d, itemID: %s, quantity: %d}", i.ItemHash, i.InstanceID, i.Quantity)
	}

	return fmt.Sprintf("Item{itemHash: %d, quantity: %d}", i.ItemHash, i.Quantity)
}

// Power is a convenience accessor to return the power level for a specific item or zero if it does not apply.
func (i *Item) Power() int {
	if i == nil || i.ItemInstance == nil || i.PrimaryStat == nil {
		return 0
	}

	return i.PrimaryStat.Value
}

// IsInVault will determine if the item is in the vault or not. True if it is; False if it is not.
func (i *Item) IsInVault() bool {
	return i.Character == nil
}

// Damage is a helper function to return the damage type of a specific item.
// NoDamage will be returned if the item does not represent a weapon which would
// not make sense to have a damage type.
func (i *Item) Damage() DamageType {
	if i == nil || i.ItemInstance == nil {
		// TODO: This should probably be kinetic but currently it represents noDamage. How should this be fixed?
		return DamageType(0)
	}

	return i.DamageType
}
