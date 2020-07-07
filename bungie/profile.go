package bungie

import (
	"errors"
	"sort"
	"time"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
)

// Profile contains all information about a specific Destiny membership, including character and
// inventory information.
type Profile struct {
	MembershipType        int
	MembershipID          string
	BungieNetMembershipID string
	DateLastPlayed        time.Time
	DisplayName           string
	Characters            CharacterList

	AllItems ItemList

	// A map between the character ID and the Loadout currently equipped on that character
	Loadouts map[string]Loadout

	Equipments map[string]Equipment

	// NOTE: Still not sure this is the best approach to flatten items into a single list,
	// it works well for now so we will go with it. There are too many potential spots to
	// look for an item.
	//Equipments       map[string]ItemList
	//Inventories      map[string]ItemList
	//ProfileInventory ItemList
	//Currencies       ItemList
}

// ProfileMsg is a wrapper around a Profile struct that should be used exclusively for sending a
// Profile over a channel, or at least in cases where an error also needs to be sent to indicate
// failures.
type ProfileMsg struct {
	*Profile
	error
}

// GetProfileForCurrentUser will retrieve the Profile data for the currently logged in user
// (determined by the access_token)
func GetProfileForCurrentUser(client *Client, requireInstanceData bool) (*Profile, error) {

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Could not load profile with access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return nil, errors.New("Couldn't load current user information")
	}

	// This will always be the Destiny membership with the most recently played character
	membership := currentAccount.DestinyMembership

	profileResponse := GetProfileResponse{}
	err := client.Execute(NewUserProfileRequest(membership.MembershipType,
		membership.MembershipID), &profileResponse)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Profile response from Bungie!: %s", err.Error())
		return nil, errors.New("Failed to read current user's profile: " + err.Error())
	}

	profile := fixupProfileFromProfileResponse(&profileResponse, false)
	profile.BungieNetMembershipID = currentAccount.BungieNetUser.MembershipID

	for _, char := range profile.Characters {
		glg.Debugf("Character(%s) last played date: %+v", classHashToName[char.ClassHash], char.DateLastPlayed)
	}

	return profile, nil
}

func fixupProfileFromProfileResponse(response *GetProfileResponse, requireInstanceData bool) *Profile {
	profile := &Profile{}

	// Profile Component
	profile.MembershipID = response.membershipID()
	profile.MembershipType = response.membershipType()

	// Transform character map into an ordered list based on played time.
	// Characters Component
	if response.Response.Characters != nil {
		profile.Characters = make([]*Character, 0, len(response.Response.Characters.Data))
		for _, char := range response.Response.Characters.Data {
			profile.Characters = append(profile.Characters, char)
		}

		sort.Sort(sort.Reverse(LastPlayedSort(profile.Characters)))
	}

	// Flatten out the items from different buckets including currencies, inventories, eequipments,
	// etc.
	//totalItemCount := len(response.Response.ProfileCurrencies.Data.Items) + len(response.Response.ProfileInventory.Data.Items)
	// for id := range response.Response.Characters.Data {
	// 	totalItemCount += len(response.Response.CharacterEquipment.Data[id].Items)
	// 	totalItemCount += len(response.Response.CharacterInventories.Data[id].Items)
	// }

	items := make(ItemList, 0, 32)

	// ProfileCurrencies Component
	if response.Response.ProfileCurrencies != nil {
		if !requireInstanceData {
			items = append(items, response.Response.ProfileCurrencies.Data.Items...)
		}
	}

	// ProfileInventory Component
	if response.Response.ProfileInventory != nil {
		for _, item := range response.Response.ProfileInventory.Data.Items {
			item.ItemInstance = response.instanceData(item.InstanceID)

			if !requireInstanceData || item.ItemInstance != nil {
				items = append(items, item)
			}
		}
	}

	// CharacterEquipment Component
	if response.Response.CharacterEquipment != nil {

		// If the character equipment fields were provided, populate the profile's loadouts map
		profile.Loadouts = make(map[string]Loadout)
		profile.Equipments = make(map[string]Equipment)

		for charID, list := range response.Response.CharacterEquipment.Data {

			currentLoadout := newLoadout()
			currentEquipment := newEquipment()

			for _, item := range list.Items {

				item.Character = response.character(charID)
				item.ItemInstance = response.instanceData(item.InstanceID)

				// We don't need to check IsEquipped here, that is what the CharacterEquipment
				// group means, just make sure its on the right character.
				if equipmentBucket, ok := EquipmentBucketLookup[item.BucketHash]; ok {
					currentLoadout[equipmentBucket] = item

					// By doing equipped items before the rest of the inventory below, this will
					// ensure that the first item in the slice is always the currently
					// equipped item for that bucket
					currentEquipment[equipmentBucket] = append(currentEquipment[equipmentBucket], item)
				}
				if !requireInstanceData || item.ItemInstance != nil {
					items = append(items, item)
				}
			}

			profile.Loadouts[charID] = currentLoadout
			profile.Equipments[charID] = currentEquipment
		}
	}

	// CharacterInventories Component
	if response.Response.CharacterInventories != nil {

		if profile.Equipments == nil {
			profile.Equipments = make(map[string]Equipment)
		}

		for charID, list := range response.Response.CharacterInventories.Data {

			currentEquipment := profile.Equipments[charID]
			if currentEquipment == nil {
				currentEquipment = newEquipment()
				profile.Equipments[charID] = currentEquipment
			}

			for _, item := range list.Items {
				item.Character = response.character(charID)
				item.ItemInstance = response.instanceData(item.InstanceID)

				if equipmentBucket, ok := EquipmentBucketLookup[item.BucketHash]; ok {

					// Since this is going after the CharacterEquipment (above),
					// the rest of the slice will be the currnet unequipped items
					currentEquipment[equipmentBucket] = append(currentEquipment[equipmentBucket], item)
				}
				if !requireInstanceData || item.ItemInstance != nil {
					items = append(items, item)
				}
			}
		}
	}

	profile.AllItems = items

	return profile
}
