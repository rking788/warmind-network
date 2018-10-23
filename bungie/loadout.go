package bungie

import (
	"math/rand"
	"sort"

	"github.com/getsentry/raven-go"
	"github.com/rking788/warmind-network/models"

	"github.com/kpango/glg"
)

// fromPersistedLoadout is responsible for searching through the Profile and
// equipping the weapons described in the PersistedLoadout. A best attempt will
// be made to equip the same instances of the gear persisted but as a fallback
// the same item hashes could be used. That way if the user deleted one
// instance of a weapon they will still get that weapon equipped if they picked
// up a new instance of one.
func fromPersistedLoadout(persisted models.PersistedLoadout, profile *models.Profile) models.Loadout {

	result := make(models.Loadout)
	for equipmentBucket, item := range persisted {
		sameHashList := ItemList(profile.AllItems).FilterItemsBubble(itemHashFilter, item.ItemHash)
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

func findMaxLightLoadout(profile *models.Profile, destinationID string) models.Loadout {

	// Start by filtering all items that are NOT exotics
	destinationCharacter := profile.Characters.FindCharacterFromID(destinationID)
	destinationClassType := destinationCharacter.ClassType
	filteredItems := ItemList(profile.AllItems).
		FilterItemsMultipleBubble(createItemClassTypeFilter(destinationClassType),
			createItemNotTierTypeFilter(exoticTier),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	gearSortedByLight := groupAndSortGear(filteredItems)

	// Find the best loadout given just legendary weapons
	loadout := make(models.Loadout)
	for i := models.Kinetic; i <= models.ClassArmor; i++ {
		if i == models.Ghost {
			continue
		}
		loadout[i] = findBestItemForBucket(i, gearSortedByLight[i], destinationID)
	}

	// Determine the best exotics to use for both weapons and armor
	exotics := ItemList(profile.AllItems).
		FilterItemsMultipleBubble(createItemTierTypeFilter(exoticTier),
			createItemClassTypeFilter(destinationClassType),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	exoticsSortedAndGrouped := groupAndSortGear(exotics)

	// Override inventory items with exotics as needed
	for _, bucket := range []models.EquipmentBucket{models.ClassArmor} {
		exoticCandidate := findBestItemForBucket(bucket, exoticsSortedAndGrouped[bucket], destinationID)
		if exoticCandidate != nil && exoticCandidate.Power() > loadout[bucket].Power() {
			loadout[bucket] = exoticCandidate
			glg.Debugf("Overriding %s...", bucket)
		}
	}

	var weaponExoticCandidate *models.Item
	var weaponBucket models.EquipmentBucket
	for _, bucket := range [3]models.EquipmentBucket{models.Kinetic, models.Energy, models.Power} {
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

	var armorExoticCandidate *models.Item
	var armorBucket models.EquipmentBucket
	for _, bucket := range [4]models.EquipmentBucket{models.Helmet, models.Gauntlets, models.Chest, models.Legs} {
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

func findRandomLoadout(profile *models.Profile, destinationID string, includeArmor bool) models.Loadout {

	destinationCharacter := profile.Characters.FindCharacterFromID(destinationID)
	destinationClassType := destinationCharacter.ClassType
	loadout := profile.Loadouts[destinationID]
	equipment := profile.Equipments[destinationID]

	glg.Debugf("Starting randomize with loadout charID(%s): %+v", destinationID, loadout)

	// Only randomize weapons unless specified
	upperBound := models.Power
	if includeArmor {
		upperBound = models.Legs
	}

	for i := models.Kinetic; i <= upperBound; i++ {
		if i == models.Ghost {
			continue
		}

		unequippedItems := ItemList(equipment[i][1:])
		filteredBucket := unequippedItems.FilterItemsMultipleBubble(createItemClassTypeFilter(destinationClassType),
			createItemNotTierTypeFilter(exoticTier),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))

		bucketCount := len(filteredBucket)
		if bucketCount <= 0 {
			continue
		}

		// This will give a number in the range [0, n-1), that way it'll be safe to do the +1
		// adjustment below. This will skip past the curently equipped item at index 0 and only
		// select a non-equipped weapon in the bucket
		randIndex := rand.Intn(bucketCount)
		loadout[i] = filteredBucket[randIndex]
	}

	// Override inventory items with exotics as needed

	// Exotic Weapon
	// Need the +1 here because rand.Intn does not include the integer arg in the range
	randExoticBucket := models.EquipmentBucket(rand.Intn(int(models.Power)) + 1)

	filteredExoticWeapons := ItemList(equipment[randExoticBucket][1:]).FilterItemsMultipleBubble(
		createItemTierTypeFilter(exoticTier),
		createItemClassTypeFilter(destinationClassType),
		createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
	if len(filteredExoticWeapons) > 0 {
		randIndex := rand.Intn(len(filteredExoticWeapons))
		loadout[randExoticBucket] = filteredExoticWeapons[randIndex]
	}

	// Exotic Armor
	if includeArmor {
		randExoticBucket = models.EquipmentBucket(rand.Intn(int(models.Legs)-int(models.Helmet)) + int(models.Helmet))
		filteredExoticArmor := ItemList(equipment[randExoticBucket][1:]).FilterItemsMultipleBubble(
			createItemTierTypeFilter(exoticTier),
			createItemClassTypeFilter(destinationClassType),
			createItemRequiredLevelFilter(destinationCharacter.LevelProgression.Level))
		if len(filteredExoticArmor) > 0 {
			randIndex := 0
			if len(filteredExoticArmor) > 1 {
				randIndex = rand.Intn(len(filteredExoticArmor))
			}
			loadout[randExoticBucket] = filteredExoticArmor[randIndex]
		}
	}

	glg.Debugf("Ending randomize with loadout charID(%s): %+v", destinationID, loadout)

	return loadout
}

func equipLoadout(loadout models.Loadout, destinationID string, profile *models.Profile, membershipType int, client *Client) error {

	characters := profile.Characters
	itemsToVault := make([]*models.Item, 0, 10)

	// Swap any items that are currently equipped on other characters to
	// prepare them to be transferred
	for bucket, item := range loadout {
		if item == nil {
			glg.Warnf("Found nil item for bucket: %v", bucket)
			continue
		}

		// Free any space in a bucket that needs to receive a new item but is currently full
		if len(profile.Equipments[destinationID][bucket]) == MaxItemsPerBucket &&
			(item.Character == nil || item.CharacterID != destinationID) {
			for i := MaxItemsPerBucket - 1; i >= 1; i-- {
				// Start from the back of the bucket and move to the front, this way it will not transfer an item
				// to the vault if it is trying to equip it in the provided loadout
				if profile.Equipments[destinationID][bucket][i] == loadout[bucket] {
					glg.Debugf("Not transferring item:%d to vault because it is going to be equipped later", loadout[bucket].ItemHash)
					continue
				}

				itemsToVault = append(itemsToVault, profile.Equipments[destinationID][bucket][i])
				break
			}
		}

		if item.TransferStatus == ItemIsEquipped && item.Character != nil &&
			item.Character.CharacterID != destinationID {
			glg.Debugf("Swapping item for characterID:%d, and destinationID:%d", item.Character.CharacterID, destinationID)
			swapEquippedItem(item, profile, bucket, membershipType, client)
		}
	}

	if len(itemsToVault) != 0 {
		glg.Debugf("Tranferring items to vault to make space for loadout equip: %+v", itemsToVault)
		transferItems(itemsToVault, nil, membershipType, -1, client)
	}

	// Move all items to the destination character
	err := moveLoadoutToCharacter(loadout, destinationID, characters, membershipType, client)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error moving loadout to destination character: %s", err.Error())
		return err
	}

	// Equip all items that were just transferred
	equipItems(loadout.ToSlice(), destinationID, membershipType, client)

	return nil
}

// swapEquippedItem is responsible for equipping a new item on a character that is
// not the destination of a transfer. This way it free up the item to be equipped
// by the desired character.
func swapEquippedItem(item *models.Item, profile *models.Profile, bucket models.EquipmentBucket, membershipType int, client *Client) {

	// TODO: Currently filtering out exotics to make it easier
	// This should be more robust. There is no guarantee the character already has an exotic
	// equipped in a different slot and this may be the only option to swap out this item.
	reverseLightSortedItems := ItemList(profile.AllItems).
		FilterItemsMultipleBubble(createCharacterIDFilter(item.CharacterID),
			createItemBucketHashFilter(item.BucketHash),
			createItemNotTierTypeFilter(exoticTier))

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

func moveLoadoutToCharacter(loadout models.Loadout, destinationID string, characters models.CharacterList, membershipType int, client *Client) error {

	transferItems(loadout.ToSlice(), characters.FindCharacterFromID(destinationID),
		membershipType, -1, client)

	return nil
}

// groupAndSortGear will return a map of ItemLists. The key of the map will be the bucket type
// of all of the items in the list. Each of the lists of items will be sorted by Light value.
func groupAndSortGear(inventory ItemList) map[models.EquipmentBucket]ItemList {

	result := make(map[models.EquipmentBucket]ItemList)

	for i := models.Kinetic; i <= models.Subclass; i++ {
		result[i] = make(ItemList, 0, 20)
	}

	for _, item := range inventory {
		meta, _ := itemMetadata[item.ItemHash]
		if meta != nil {
			bkt := EquipmentBucketLookup[meta.BucketHash]
			if bkt != 0 {
				result[bkt] = append(result[bkt], item)
			}
		}
	}

	for i := models.Kinetic; i <= models.Subclass; i++ {
		sort.Sort(sort.Reverse(LightSort(result[i])))
	}

	return result
}

func sortGearBucket(bucketHash uint, inventory ItemList) ItemList {

	result := inventory.FilterItemsBubble(itemBucketHashFilter, bucketHash)
	sort.Sort(sort.Reverse(LightSort(result)))
	return result
}

func findBestItemForBucket(bucket models.EquipmentBucket, items []*models.Item, destinationID string) *models.Item {

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
