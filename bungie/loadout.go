package bungie

import (
	"math/rand"
	"sort"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
)

// Loadout will hold all currently equipped items indexed by the equipment bucket for
// which they are equipped.
type Loadout map[EquipmentBucket]*Item

func newLoadout() Loadout {
	return make(map[EquipmentBucket]*Item)
}

// Equipment is very similar to the Loadout type but contains all items in an
// equipment bucket instead of just the currently equipped item
type Equipment map[EquipmentBucket]ItemList

func newEquipment() Equipment {
	e := make(Equipment)
	for bkt := range BucketHashLookup {
		// Equipment buckets should really only have 10 items total, at least that is
		// how character inventories/equipment slots work at the time of this writing
		e[bkt] = make(ItemList, 0, 10)
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

func (l Loadout) calculateLightLevel() float64 {

	light := 0.0

	light += float64(l[Kinetic].Power()) * 0.143
	light += float64(l[Energy].Power()) * 0.143
	light += float64(l[Power].Power()) * 0.143
	// Ghosts no longer have a light/power level

	light += float64(l[Helmet].Power()) * 0.119
	light += float64(l[Gauntlets].Power()) * 0.119
	light += float64(l[Chest].Power()) * 0.119
	light += float64(l[Legs].Power()) * 0.119
	light += float64(l[ClassArmor].Power()) * 0.095

	return light
}

func (l Loadout) toSlice() []*Item {

	result := make([]*Item, 0, ClassArmor-Kinetic)
	for i := Kinetic; i <= ClassArmor; i++ {
		item := l[i]
		if item != nil {
			result = append(result, l[i])
		}
	}

	return result
}

func (l Loadout) toPersistedLoadout() PersistedLoadout {

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

// fromPersistedLoadout is responsible for searching through the Profile and
// equipping the weapons described in the PersistedLoadout. A best attempt will
// be made to equip the same instances of the gear persisted but as a fallback
// the same item hashes could be used. That way if the user deleted one
// instance of a weapon they will still get that weapon equipped if they picked
// up a new instance of one.
func fromPersistedLoadout(persisted PersistedLoadout, profile *Profile) Loadout {

	result := make(Loadout)
	for equipmentBucket, item := range persisted {
		sameHashList := profile.AllItems.FilterItemsBubble(itemHashFilter, item.ItemHash)
		if len(sameHashList) <= 0 {
			glg.Warnf("Item(%v) not in profile when restoring loadout", item.ItemHash)
			result[equipmentBucket] = nil
			continue
		}

		bestMatchItem := sameHashList[0]
		exactInstances := sameHashList.FilterItemsBubble(itemInstanceIDFilter, item.InstanceID)

		if len(exactInstances) > 0 {
			bestMatchItem = exactInstances[0]
		} else {
			glg.Warnf("Items sharing an instanceID(%v) in persisted loadout", item.ItemHash)
		}

		result[equipmentBucket] = bestMatchItem
	}

	return result
}

func findMaxLightLoadout(profile *Profile, destinationID string) Loadout {
	// Start by filtering all items that are NOT exotics
	destinationCharacter := profile.Characters.findCharacterFromID(destinationID)
	destinationClassType := destinationCharacter.ClassType
	filteredItems := profile.AllItems.
		FilterItemsMultipleBubble(createItemClassTypeFilter(destinationClassType),
			createItemNotTierTypeFilter(ExoticTier),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	gearSortedByLight := groupAndSortGear(filteredItems)

	// Find the best loadout given just legendary weapons
	loadout := make(Loadout)
	for i := Kinetic; i <= ClassArmor; i++ {
		if i == Ghost {
			continue
		}
		loadout[i] = findBestItemForBucket(i, gearSortedByLight[i], destinationID)
	}

	// Determine the best exotics to use for both weapons and armor
	exotics := profile.AllItems.
		FilterItemsMultipleBubble(createItemTierTypeFilter(ExoticTier),
			createItemClassTypeFilter(destinationClassType),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	exoticsSortedAndGrouped := groupAndSortGear(exotics)

	// Override inventory items with exotics as needed
	for _, bucket := range []EquipmentBucket{ClassArmor} {
		exoticCandidate := findBestItemForBucket(bucket, exoticsSortedAndGrouped[bucket], destinationID)
		if exoticCandidate != nil && exoticCandidate.Power() > loadout[bucket].Power() {
			loadout[bucket] = exoticCandidate
			glg.Debugf("Overriding %s...", bucket)
		}
	}

	var weaponExoticCandidate *Item
	var weaponBucket EquipmentBucket
	for _, bucket := range [3]EquipmentBucket{Kinetic, Energy, Power} {
		exoticCandidate := findBestItemForBucket(bucket, exoticsSortedAndGrouped[bucket], destinationID)
		if exoticCandidate != nil && exoticCandidate.Power() > loadout[bucket].Power() {
			if weaponExoticCandidate == nil || exoticCandidate.Power() > weaponExoticCandidate.Power() {
				weaponExoticCandidate = exoticCandidate
				weaponBucket = bucket
				glg.Debugf("Overriding %s...", bucket)
			}
		}
	}
	if weaponExoticCandidate != nil {
		loadout[weaponBucket] = weaponExoticCandidate
	}

	var armorExoticCandidate *Item
	var armorBucket EquipmentBucket
	for _, bucket := range [4]EquipmentBucket{Helmet, Gauntlets, Chest, Legs} {
		exoticCandidate := findBestItemForBucket(bucket, exoticsSortedAndGrouped[bucket], destinationID)
		if exoticCandidate != nil && exoticCandidate.Power() > loadout[bucket].Power() {
			if armorExoticCandidate == nil || exoticCandidate.Power() > armorExoticCandidate.Power() {
				armorExoticCandidate = exoticCandidate
				armorBucket = bucket
				glg.Debugf("Overriding %s...", bucket)
			}
		}
	}
	if armorExoticCandidate != nil {
		loadout[armorBucket] = armorExoticCandidate
	}

	return loadout
}

func findRandomLoadout(profile *Profile, destinationID string, includeArmor bool) Loadout {

	// Start by filtering all items that are NOT exotics
	destinationCharacter := profile.Characters.findCharacterFromID(destinationID)
	destinationClassType := destinationCharacter.ClassType
	filteredItems := profile.AllItems.
		FilterItemsMultipleBubble(createItemClassTypeFilter(destinationClassType),
			createItemNotTierTypeFilter(ExoticTier),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	gearSortedByLight := groupAndSortGear(filteredItems)

	// Find the best loadout given just legendary weapons
	loadout := profile.Loadouts[destinationID]
	glg.Debugf("Starting with loadout as base Loadout for charID(%s): %+v", destinationID, loadout)

	// Only randomize weapons unless specified
	upperBound := Power
	if includeArmor {
		upperBound = Legs
	}

	for i := Kinetic; i <= upperBound; i++ {
		if i == Ghost {
			continue
		}

		bucketCount := len(gearSortedByLight[i])
		if bucketCount <= 0 {
			continue
		}

		randIndex := rand.Intn(bucketCount)
		loadout[i] = gearSortedByLight[i][randIndex]
	}

	// Determine the best exotics to use for both weapons and armor
	exotics := profile.AllItems.
		FilterItemsMultipleBubble(createItemTierTypeFilter(ExoticTier),
			createItemClassTypeFilter(destinationClassType),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	exoticsSortedAndGrouped := groupAndSortGear(exotics)

	// Override inventory items with exotics as needed

	// Exotic Weapon
	randExoticBucket := EquipmentBucket(rand.Intn(int(Power)))
	if len(exoticsSortedAndGrouped[randExoticBucket]) > 0 {
		weaponBucketItems := exoticsSortedAndGrouped[randExoticBucket]
		randIndex := rand.Intn(len(weaponBucketItems))
		loadout[randExoticBucket] = weaponBucketItems[randIndex]
	}

	// Exotic Armor
	if includeArmor {
		randExoticBucket = EquipmentBucket(rand.Intn(int(Legs)-int(Helmet)) + int(Helmet))
		if len(exoticsSortedAndGrouped[randExoticBucket]) > 0 {
			armorBucketItems := exoticsSortedAndGrouped[randExoticBucket]
			randIndex := rand.Intn(len(armorBucketItems))
			loadout[randExoticBucket] = armorBucketItems[randIndex]
		}
	}

	return loadout
}

func equipLoadout(loadout Loadout, destinationID string, profile *Profile, membershipType int, client *Client) error {

	characters := profile.Characters
	// Swap any items that are currently equipped on other characters to
	// prepare them to be transferred
	for bucket, item := range loadout {
		if item == nil {
			glg.Warnf("Found nil item for bucket: %v", bucket)
			continue
		}
		if item.TransferStatus == ItemIsEquipped && item.Character != nil &&
			item.Character.CharacterID != destinationID {
			swapEquippedItem(item, profile, bucket, membershipType, client)
		}
	}

	// Move all items to the destination character
	err := moveLoadoutToCharacter(loadout, destinationID, characters, membershipType, client)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error moving loadout to destination character: %s", err.Error())
		return err
	}

	// Equip all items that were just transferred
	equipItems(loadout.toSlice(), destinationID, membershipType, client)

	return nil
}

// swapEquippedItem is responsible for equipping a new item on a character that is
// not the destination of a transfer. This way it free up the item to be equipped
// by the desired character.
func swapEquippedItem(item *Item, profile *Profile, bucket EquipmentBucket, membershipType int, client *Client) {

	// TODO: Currently filtering out exotics to make it easier
	// This should be more robust. There is no guarantee the character already has an exotic
	// equipped in a different slot and this may be the only option to swap out this item.
	reverseLightSortedItems := profile.AllItems.
		FilterItemsMultipleBubble(createCharacterIDFilter(item.CharacterID),
			createItemBucketHashFilter(item.BucketHash),
			createItemNotTierTypeFilter(ExoticTier))

	if len(reverseLightSortedItems) <= 1 {
		// TODO: If there are no other items from the specified character, then we need to
		// figure out an item to be transferred from the vault
		glg.Warn("No other items on the specified character, not currently setup to transfer new choices from the vault...")
		return
	}

	// Lowest light to highest
	sort.Sort(LightSort(reverseLightSortedItems))

	// Now that items are sorted in reverse light order, we want to equip the first
	// item in the slice, the highest light item will be the last item in the slice.
	itemToEquip := reverseLightSortedItems[0]
	character := item.Character
	equipItem(itemToEquip, character, membershipType, client)
}

func moveLoadoutToCharacter(loadout Loadout, destinationID string, characters CharacterList, membershipType int, client *Client) error {

	transferItems(loadout.toSlice(), characters.findCharacterFromID(destinationID),
		membershipType, -1, client)

	return nil
}

// groupAndSortGear will return a map of ItemLists. The key of the map will be the bucket type
// of all of the items in the list. Each of the lists of items will be sorted by Light value.
func groupAndSortGear(inventory ItemList) map[EquipmentBucket]ItemList {

	result := make(map[EquipmentBucket]ItemList)

	for i := Kinetic; i <= Subclass; i++ {
		result[i] = make(ItemList, 0, 20)
	}

	for _, item := range inventory {
		bkt := EquipmentBucketLookup[item.BucketHash]
		if bkt != 0 {
			result[bkt] = append(result[bkt], item)
		}
	}

	for i := Kinetic; i <= Subclass; i++ {
		sort.Sort(sort.Reverse(LightSort(result[i])))
	}

	return result
}

func sortGearBucket(bucketHash uint, inventory ItemList) ItemList {

	result := inventory.FilterItemsBubble(itemBucketHashFilter, bucketHash)
	sort.Sort(sort.Reverse(LightSort(result)))
	return result
}

func findBestItemForBucket(bucket EquipmentBucket, items []*Item, destinationID string) *Item {

	if len(items) <= 0 {
		return nil
	}

	candidate := items[0]
	for i := 1; i < len(items); i++ {
		next := items[i]
		if next.Power() < candidate.Power() {
			// Lower light value, keep the current candidate
			break
		} else if candidate.IsEquipped && candidate.CharacterID == destinationID {
			// The current max light piece of gear is currently equipped on the
			// destination character, avoiding moving items around if we don't need to.
			break
		}

		if (!next.IsInVault() && next.CharacterID == destinationID) &&
			(!candidate.IsInVault() && candidate.CharacterID != destinationID) {
			// This next item is the same light and on the destination character already,
			// the current candidate is not
			candidate = next
		} else if (!next.IsInVault() && next.CharacterID == destinationID) &&
			(!candidate.IsInVault() && candidate.CharacterID == destinationID) {
			if next.TransferStatus == ItemIsEquipped && candidate.TransferStatus != ItemIsEquipped {
				// The next item is currently equipped on the destination character,
				// the current candidate is not
				candidate = next
			}
		} else if (!candidate.IsInVault() && candidate.CharacterID != destinationID) &&
			next.IsInVault() {
			// If the current candidate is on a character that is NOT the destination and the next
			// candidate is in the vault, prefer that since we will only need to do a
			// single transfer request
			candidate = next
		}
	}

	return candidate
}
