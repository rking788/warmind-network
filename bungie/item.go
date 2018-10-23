package bungie

import (
	"github.com/kpango/glg"
	"github.com/rking788/warmind-network/models"
)

// ItemList is a collection of Item instances.
type ItemList []*models.Item

// ItemFilter is a type that will be used as a paramter to a filter function.
// The parameter will be a function pointer. The function pointed to will need to return
// true if the element meets some criteria and false otherwise. If the result of
// this filter is false, then the item will be removed.
type ItemFilter func(*models.Item, interface{}) bool

/*
 * Sort Conformance Methods
 */

// LightSort specifies a specific type for ItemList that can be sorted by Light value of each item.
type LightSort ItemList

func (items LightSort) Len() int      { return len(items) }
func (items LightSort) Swap(i, j int) { items[i], items[j] = items[j], items[i] }
func (items LightSort) Less(i, j int) bool {
	return items[i].Power() < items[j].Power()
}

// FilterItems will filter the receiver slice of Items and return only the items that match
// the criteria specified in ItemFilter. If ItemFilter returns True, the element will be included,
// if it returns False the element will be removed.
func (items ItemList) _FilterItems(filter ItemFilter, arg interface{}) ItemList {

	result := make(ItemList, 0, len(items))

	for _, item := range items {
		if filter(item, arg) {
			result = append(result, item)
		}
	}

	return result
}

func (items ItemList) _FilterItemsMultiple(filters ...ItemFilter) ItemList {
	result := make(ItemList, 0, len(items))

	for _, item := range items {
		allMatch := true

		for _, filter := range filters {
			allMatch = allMatch && filter(item, nil)
			if allMatch == false {
				break
			}
		}

		if allMatch {
			result = append(result, item)
		}
	}

	return result
}

// FilterItemsBubble will use the "bubble" approach to filter items in the provided list using the
// provided filter method and argument.
func (items ItemList) FilterItemsBubble(filter ItemFilter, arg interface{}) ItemList {

	filterPos := 0

	for i, item := range items {
		if filter(item, arg) {
			tempItem := items[filterPos]
			items[filterPos] = items[i]
			items[i] = tempItem
			filterPos++
		}
	}

	result := make(ItemList, filterPos)
	copy(result, items)

	return result
}

// FilterItemsMultipleBubble will use the bubble approach and provided filters to filter the items
// in the item given item list.
func (items ItemList) FilterItemsMultipleBubble(filters ...ItemFilter) ItemList {

	filterPos := 0

	for i, item := range items {
		allMatch := true

		for _, filter := range filters {
			allMatch = allMatch && filter(item, nil)
			if allMatch == false {
				break
			}
		}

		if allMatch {
			tempItem := items[filterPos]
			items[filterPos] = items[i]
			items[i] = tempItem
			filterPos++
		}
	}

	result := make(ItemList, filterPos)
	copy(result, items)

	return result
}

// itemHashFilter will return true if the itemHash provided matches the hash of the item;
// otherwise false.
func itemHashFilter(item *models.Item, itemHash interface{}) bool {
	return item != nil && (item.ItemHash == itemHash.(uint))
}

// itemHashesFilter will return true if the item's hash value is present in the provided
// slice of hashes; otherwise false.
func itemHashesFilter(item *models.Item, hashList interface{}) bool {
	for _, hash := range hashList.([]uint) {
		return itemHashFilter(item, hash)
	}

	return false
}

func createItemBucketHashFilter(bucketTypeHash interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, unused interface{}) bool {
		return itemBucketHashFilter(item, bucketTypeHash)
	}
}

// itemBucketHashIncludingVaultFilter will filter the list of items by the specified
// bucket hash or the Vault location
func itemBucketHashFilter(item *models.Item, bucketTypeHash interface{}) bool {

	if metadata, ok := itemMetadata[item.ItemHash]; ok {
		return metadata.BucketHash == bucketTypeHash.(uint)
	}

	glg.Warnf("No metadata found for item: %d", item.ItemHash)
	return false
}

func createCharacterIDFilter(characterID interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, unused interface{}) bool {
		return itemCharacterIDFilter(item, characterID)
	}
}

// itemCharacterIDFilter will filter the list of items by the specified character identifier
func itemCharacterIDFilter(item *models.Item, characterID interface{}) bool {
	return item.Character != nil && (item.Character.CharacterID == characterID.(string))
}

// itemIsEngramFilter will return true if the item represents an engram; otherwise false.
func itemIsEngramFilter(item *models.Item, wantEngram interface{}) bool {
	_, isEngram := engramHashes[item.ItemHash]
	return isEngram == wantEngram.(bool)
}

// itemTierTypeFilter is a filter that will filter out items that are not of the specified tier.
func itemTierTypeFilter(item *models.Item, tierType interface{}) bool {
	metadata, ok := itemMetadata[item.ItemHash]
	if !ok {
		glg.Warnf("No metadata found for item: %d", item.ItemHash)
		return false
	}
	return metadata.TierType == tierType.(int)
}

func createItemTierTypeFilter(tierType interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, tier interface{}) bool {
		return itemTierTypeFilter(item, tierType)
	}
}

func createItemNotTierTypeFilter(tierType interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, tier interface{}) bool {
		return itemNotTierTypeFilter(item, tierType)
	}
}

func itemNotTierTypeFilter(item *models.Item, tierType interface{}) bool {

	if metadata, ok := itemMetadata[item.ItemHash]; ok {
		return metadata.TierType != tierType.(int)
	}

	glg.Warnf("No metadata found for item: %s", item.ItemHash)
	return false
}

// itemInstanceIDFilter is an item filter that will return true for all items with an
// instanceID property equal to the one provided. This is useful for filtering a list
// down to a specific instance of an item.
func itemInstanceIDFilter(item *models.Item, instanceID interface{}) bool {
	return item.InstanceID == instanceID.(string)
}

func createItemClassTypeFilter(classType interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, class interface{}) bool {
		return itemClassTypeFilter(item, classType)
	}
}

// itemClassTypeFilter will filter out all items that are not equippable by the specified class
func itemClassTypeFilter(item *models.Item, classType interface{}) bool {

	if metadata, ok := itemMetadata[item.ItemHash]; ok {
		return (metadata.ClassType == UnknownClassType) ||
			(metadata.ClassType == classType.(int))
	}

	glg.Warnf("No metadata found for item: %s", item.ItemHash)
	return false
}

func createItemRequiredLevelFilter(maxLevel interface{}) func(*models.Item, interface{}) bool {

	return func(item *models.Item, level interface{}) bool {
		return itemRequiredLevelFilter(item, maxLevel)
	}
}

// itemRequiredLevelFilter will filter items that have a required level that is greater than
// the level provided in `maxLevel`. True if the required level is <= the max level; False otherwise
func itemRequiredLevelFilter(item *models.Item, maxLevel interface{}) bool {
	return item.ItemInstance != nil && item.ItemInstance.EquipRequiredLevel <= maxLevel.(int)
}
