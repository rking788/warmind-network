package bungie

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/kpango/glg"
	"github.com/rking788/warmind-network/db"
)

const (
	// TransferDelay will be the artificial between transfer requests to try and avoid throttling
	TransferDelay = 750 * time.Millisecond
)

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

// Clients stores a list of bungie.Client instances that can be used to make HTTP requests to the Bungie API
var Clients *ClientPool

// It is probably faster to just load all of the item_name->item_hash lookups into memory.
// That way we can give feedback to the user quicker if an item name is not found.
// If memory overhead becomes an issue this can be removed and go back to the DB lookups.
var itemHashLookup map[string]uint

var engramHashes map[uint]bool
var itemMetadata map[uint]*ItemMetadata
var bungieAPIKey string
var warmindAPIKey string

// BucketHashLookup maps all of the equipment bucket constants to their corresponding bucket
// hashes as defined in the Bungie API.
var BucketHashLookup map[EquipmentBucket]uint

// EquipmentBucketLookup maps the bucket hash values defined in the Bungie API to the bucket
// equipment constants.
var EquipmentBucketLookup map[uint]EquipmentBucket

// RecordLookup will hold all of the record information to be used for character
// titles related to equipped titles from seals.
var recordLookup map[uint]*Record

// InitEnv provides a package level initialization point for any work that is environment specific
func InitEnv(apiKey, warmindKey string) {
	bungieAPIKey = apiKey
	warmindAPIKey = warmindKey

	Clients = NewClientPool()

	err := PopulateEngramHashes()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error populating engram hashes: %s\nExiting...", err.Error())
		return
	}
	err = PopulateBucketHashLookup()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error populating bucket hash values: %s\nExiting...", err.Error())
		return
	}
	err = PopulateItemMetadata()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error populating item metadata lookup table: %s\nExiting...", err.Error())
		return
	}

	err = populateRecordLookup()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error populating record lookup table: %s\nExiting...", err.Error())
		return
	}
}

// EquipmentBucket is the type of the key for the bucket type hash lookup
type EquipmentBucket uint

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

// PopulateEngramHashes will intialize the map holding all item_hash values that represent engram types.
func PopulateEngramHashes() error {

	var err error
	engramHashes, err = db.FindEngramHashes()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error populating engram item_hash values: %s", err.Error())
		return err
	} else if len(engramHashes) <= 0 {
		raven.CaptureError(err, nil)
		glg.Error("Didn't find any engram item hashes in the database.")
		return errors.New("No engram item_hash values found")
	}

	glg.Infof("Loaded %d hashes representing engrams into the map.", len(engramHashes))
	return nil
}

// PopulateItemMetadata is responsible for loading all of the metadata fields that need
// to be loaded into memory for common inventory related operations.
func PopulateItemMetadata() error {

	rows, err := db.LoadItemMetadata()
	if err != nil {
		return err
	}
	defer rows.Close()

	itemMetadata = make(map[uint]*ItemMetadata)
	itemHashLookup = make(map[string]uint)
	for rows.Next() {

		var hash, itemType uint
		var itemName string
		itemMeta := ItemMetadata{}
		rows.Scan(&hash, &itemName, &itemMeta.TierType, &itemMeta.ClassType, &itemMeta.BucketHash, &itemType)
		if itemType == Dummy || itemType == None {
			continue
		}

		itemMetadata[hash] = &itemMeta
		if itemName != "" {
			itemHashLookup[itemName] = hash
		} else {
			glg.Debug("Found an empty item name, skipping...")
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	glg.Infof("Loaded %d item metadata entries", len(itemMetadata))

	return nil
}

func populateRecordLookup() error {

	rows, err := db.LoadRecords()
	if err != nil {
		return err
	}
	defer rows.Close()
	recordLookup = make(map[uint]*Record)

	for rows.Next() {

		r := Record{}
		rows.Scan(&r.Hash, &r.Name, &r.HasTitle, &r.MaleTitle, &r.FemaleTitle)

		recordLookup[r.Hash] = &r
	}

	glg.Infof("Loaded %d record entries", len(recordLookup))

	return nil
}

func ItemNames() []string {
	result := make([]string, 0, len(itemMetadata))

	for name := range itemHashLookup {
		result = append(result, name)
	}

	return result
}

// PopulateBucketHashLookup will fill the map that will be used to lookup bucket type hashes
// which will be used to determine which type of equipment a specific Item represents.
// From the DestinyInventoryBucketDefinition table in the manifest.
func PopulateBucketHashLookup() error {

	// TODO: This absolutely needs to be done dynamically from the manifest. Not from a
	//static definition
	BucketHashLookup = make(map[EquipmentBucket]uint)

	BucketHashLookup[Kinetic] = 1498876634
	BucketHashLookup[Energy] = 2465295065
	BucketHashLookup[Power] = 953998645
	BucketHashLookup[Ghost] = 4023194814

	BucketHashLookup[Helmet] = 3448274439
	BucketHashLookup[Gauntlets] = 3551918588
	BucketHashLookup[Chest] = 14239492
	BucketHashLookup[Legs] = 20886954
	BucketHashLookup[Artifact] = 434908299
	BucketHashLookup[ClassArmor] = 1585787867
	BucketHashLookup[Subclass] = 3284755031

	EquipmentBucketLookup = make(map[uint]EquipmentBucket)
	EquipmentBucketLookup[1498876634] = Kinetic
	EquipmentBucketLookup[2465295065] = Energy
	EquipmentBucketLookup[953998645] = Power
	EquipmentBucketLookup[4023194814] = Ghost

	EquipmentBucketLookup[3448274439] = Helmet
	EquipmentBucketLookup[3551918588] = Gauntlets
	EquipmentBucketLookup[14239492] = Chest
	EquipmentBucketLookup[20886954] = Legs
	EquipmentBucketLookup[434908299] = Artifact
	EquipmentBucketLookup[1585787867] = ClassArmor
	EquipmentBucketLookup[3284755031] = Subclass

	return nil
}

// CountItem will count the number of the specified item
func CountItem(itemName, accessToken string) (string, error) {
	glg.Infof("ItemName: %s", itemName)

	// Check common misinterpretations from Alexa
	if translation, ok := commonAlexaItemTranslations[itemName]; ok {
		itemName = translation
	}

	hash, ok := itemHashLookup[itemName]
	if !ok {
		glg.Warnf("Hash lookup failed for item with name=%s", itemName)
		outputStr := fmt.Sprintf("Sorry Guardian, I could not find any items named %s in your inventory.", itemName)
		return outputStr, nil
	}

	glg.Infof("ItemHash: %d", hash)
	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	// Load all items on all characters
	profile, err := GetProfileForCurrentUser(client, false, false)
	if err != nil {
		outputStr := "Sorry Guardian, I could not load your items from Destiny, you may need to re-link your account in the Alexa app."
		return outputStr, nil
	}

	matchingItems := profile.AllItems.FilterItemsBubble(itemHashFilter, hash)
	glg.Infof("Found %d items entries in characters inventory.", len(matchingItems))

	if len(matchingItems) == 0 {
		outputStr := fmt.Sprintf("You don't have any %s on any of your characters.", itemName)
		return outputStr, nil
	}

	outputString := ""
	for _, item := range matchingItems {
		if item.Character == nil {
			outputString += fmt.Sprintf("You have %d %s on your account", item.Quantity, itemName)
		} else {
			outputString += fmt.Sprintf("Your %s has %d %s. ", classHashToName[item.Character.ClassHash], item.Quantity, itemName)
		}
	}

	return outputString, nil
}

// TransferItem is responsible for calling the necessary Bungie.net APIs to
// transfer the specified item to the specified character. The quantity is optional
// as well as the source class. If no quantity is specified, all of the specific
// items will be transfered to the particular character.
func TransferItem(itemName, accessToken, sourceClass, destinationClass string, count int) (string, error) {
	glg.Infof("ItemName: %s, Source: %s, Destination: %s, Count: %d", itemName, sourceClass, destinationClass, count)

	// Check common misinterpretations from Alexa
	if translation, ok := commonAlexaItemTranslations[itemName]; ok {
		itemName = translation
	}
	if translation, ok := commonAlexaClassNameTrnaslations[destinationClass]; ok {
		destinationClass = translation
	}
	if translation, ok := commonAlexaClassNameTrnaslations[sourceClass]; ok {
		sourceClass = translation
	}

	hash, ok := itemHashLookup[itemName]
	if !ok {
		outputStr := fmt.Sprintf("Sorry Guardian, I could not find any items named %s in your inventory.", itemName)
		return outputStr, nil
	}

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	profile, err := GetProfileForCurrentUser(client, false, true)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Items response from Bungie!: %s", err.Error())
		return "", err
	}

	matchingItems := profile.AllItems.FilterItemsBubble(itemHashFilter, hash)
	glg.Infof("Found %d items entries in characters inventory.", len(matchingItems))

	if len(matchingItems) == 0 {
		outputStr := fmt.Sprintf("You don't have any %s on any of your characters.", itemName)
		return outputStr, nil
	}

	destCharacter, err := profile.Characters.findDestinationCharacter(destinationClass)
	if err != nil {
		output := fmt.Sprintf("Sorry Guardian, I could not transfer your %s because you do not have any %s characters in Destiny.", itemName, destinationClass)
		raven.CaptureError(err, nil)
		glg.Error(output)

		db.InsertUnknownValueIntoTable(destinationClass, db.UnknownClassTable)
		return output, nil
	}

	actualQuantity := transferItems(matchingItems, destCharacter, profile.MembershipType,
		count, client)

	var output string
	if count != -1 && actualQuantity < count {
		output = fmt.Sprintf("You only had %d %s on other characters, all of it has been transferred to your %s", actualQuantity, itemName, destinationClass)
	} else {
		output = fmt.Sprintf("All set Guardian, %d %s have been transferred to your %s", actualQuantity, itemName, destinationClass)
	}

	return output, nil
}

// EquipMaxLightGear will equip all items that are required to have the maximum light on a character
func EquipMaxLightGear(accessToken string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	profile, err := GetProfileForCurrentUser(client, true, true)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Items response from Bungie!: %s", err.Error())
		return "", err
	}

	// Transfer to the most recent character on the most recent platform
	destinationCharacter := profile.Characters[0]
	membershipType := profile.MembershipType

	glg.Debugf("Character(%s), MembershipID(%s), MembershipType(%d)",
		profile.Characters[0].CharacterID, profile.MembershipID, profile.MembershipType)

	loadout := findMaxLightLoadout(profile, destinationCharacter.CharacterID)

	glg.Debugf("Found loadout to equip: %v", loadout)
	glg.Infof("Calculated power for loadout: %f", loadout.calculateLightLevel())

	err = equipLoadout(loadout, destinationCharacter.CharacterID, profile, membershipType, client)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to equip the specified loadout: %s", err.Error())
		return "", err
	}

	characterClass := classHashToName[destinationCharacter.ClassHash]
	titleHash := destinationCharacter.TitleRecordHash
	title := "Guardian"
	if titleHash != 0 {
		if record, ok := recordLookup[titleHash]; ok {
			if destinationCharacter.GenderHash == MALE {
				title = record.MaleTitle
			} else if destinationCharacter.GenderHash == FEMALE {
				title = record.FemaleTitle
			} else {
				glg.Warnf("Unknown character gender hash: %d")
			}
		} else {
			glg.Warnf("Title record not found with hash=%d", titleHash)
		}
	}
	outputStr := fmt.Sprintf("Max power equipped to your %s %s. You are a force "+
		"to be wreckoned with.", characterClass, title)
	return outputStr, nil
}

// RandomizeLoadout will determine random items to be equipped on the current character.
// It is possible to ask for only weapons to be random or for weapons & armor to be random.
func RandomizeLoadout(accessToken string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	profile, err := GetProfileForCurrentUser(client, true, true)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Items response from Bungie!: %s", err.Error())
		return "", err
	}

	// Transfer to the most recent character on the most recent platform
	destinationID := profile.Characters[0].CharacterID
	membershipType := profile.MembershipType

	glg.Debugf("Character(%s), MembershipID(%s), MembershipType(%d)",
		profile.Characters[0].CharacterID, profile.MembershipID, profile.MembershipType)

	randomLoadout := findRandomLoadout(profile, destinationID, false)

	glg.Debugf("Found random loadout to equip: %v", randomLoadout)
	glg.Infof("Calculated power for loadout: %f", randomLoadout.calculateLightLevel())

	err = equipLoadout(randomLoadout, destinationID, profile, membershipType, client)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to equip the specified loadout: %s", err.Error())
		return "", err
	}

	characterClass := classHashToName[profile.Characters[0].ClassHash]
	outputStr := fmt.Sprintf("Random gear equipped to your %s Guardian. Goodluck "+
		"with that.", characterClass)
	return outputStr, nil
}

// UnloadEngrams is responsible for transferring all engrams off of a character and
func UnloadEngrams(accessToken string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	profile, err := GetProfileForCurrentUser(client, false, false)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Items response from Bungie!: %s", err.Error())
		return "", err
	}

	matchingItems := profile.AllItems.FilterItemsBubble(itemIsEngramFilter, true)
	if len(matchingItems) == 0 {
		outputStr := fmt.Sprintf("You don't have any engrams on your current character. " +
			"Happy farming Guardian!")
		return outputStr, nil
	}

	foundCount := 0
	for _, item := range matchingItems {
		foundCount += item.Quantity
	}

	glg.Infof("Found %d engrams on all characters", foundCount)

	_ = transferItems(matchingItems, nil, profile.MembershipType, -1, client)

	var output string
	output = fmt.Sprintf("All set Guardian, your engrams have been transferred to your vault. " +
		"Happy farming Guardian")

	return output, nil
}

// DoesLoadoutExist will check to see if a loadout exists for the current user with the given name.
// This should be done before creating a loadout to verify with the user whether the would
// like to overwrite the existing one or not.
func DoesLoadoutExist(accessToken, name string) (bool, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Could not load current profile with access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return false, errors.New("Couldn't load the current account")
	}

	// Check to see if a loadout with this name already exists and prompt for
	// confirmation to overwrite
	bnetMembershipID := currentAccount.BungieNetUser.MembershipID

	existing, _ := db.SelectLoadout(bnetMembershipID, name)
	return existing != "", nil
}

// CreateLoadoutForCurrentCharacter will create a new PersistedLoadout based on the items equipped
// to the user's current character and save them to the persistent storage.
func CreateLoadoutForCurrentCharacter(accessToken, name string, shouldOverwrite bool) (string, error) {

	glg.Infof("Creating new loadout named: %s", name)
	if name == "" {
		outputStr := "Sorry Guardian, you need to provide a loadout name based on an " +
			"activtiy like crucible, strikes, or patrols."
		return outputStr, nil
	}

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Could not load current profile with access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return "", errors.New("Couldn't load the current account")
	}

	// Check to see if a loadout with this name already exists and prompt for
	// confirmation to overwrite
	bnetMembershipID := currentAccount.BungieNetUser.MembershipID
	exists, err := DoesLoadoutExist(accessToken, name)
	if !shouldOverwrite && (err != nil || exists) {
		// Prompt the user to see if they want to overwrite the existing loadout
		outputStr := fmt.Sprintf("You already have a loadout named %s, it has not been overwritten.",
			name)
		return outputStr, nil
	}

	// This will be the membership that has the most recently played character
	membership := currentAccount.DestinyMembership

	profileResponse := GetProfileResponse{}
	err = client.Execute(NewGetCurrentEquipmentRequest(membership.MembershipType,
		membership.MembershipID), &profileResponse)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Profile response from Bungie!: %s", err.Error())
		return "", errors.New("Failed to read current user's profile: " + err.Error())
	}

	// At some point it could be useful to save emotes, ships, subclasses, etc. That is why
	// instance data is not required here for getting the profile info.
	// TOOD: Maybe requiring it to be equippable makes sense here? the one thing
	// i am stuck on is subclasses. Are subclasses equippable?
	profile := fixupProfileFromProfileResponse(&profileResponse, false, false)
	profile.BungieNetMembershipID = bnetMembershipID

	loadout := profile.Loadouts[profile.Characters[0].CharacterID]

	glg.Debugf("Created Loadout: %+v", loadout)

	persistedLoadout := loadout.toPersistedLoadout()
	persistedBytes, err := json.Marshal(persistedLoadout)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to marshal the loadout to JSON: %s", err.Error())
		return "", err
	}

	if shouldOverwrite {
		// The user has confirmed they want to overwrite the existing loadout
		db.UpdateLoadout(persistedBytes, bnetMembershipID, name)
	} else {
		// The user does not have a loadout with this name so save a new one
		db.SaveLoadout(persistedBytes, bnetMembershipID, name)
	}

	outputStr := "All set Guardian, your " + name + " loadout was saved for you."
	return outputStr, nil
}

// EquipNamedLoadout will read the loadout with the specified name from the persistent
// store and then equip it on the user's current character.
func EquipNamedLoadout(accessToken, name string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Could not load current profile with access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return "", errors.New("Couldn't load the current account")
	}

	// This will always be the Destiny membership with the most recently played character
	membership := currentAccount.DestinyMembership

	profileResponse := GetProfileResponse{}
	err := client.Execute(NewUserProfileRequest(membership.MembershipType,
		membership.MembershipID), &profileResponse)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Profile response from Bungie!: %s", err.Error())
		return "", errors.New("Failed to read current user's profile: " + err.Error())
	}

	// TOOD: Maybe requiring it to be equippable makes sense here? the one thing
	// i am stuck on is subclasses. Are subclasses equippable?
	profile := fixupProfileFromProfileResponse(&profileResponse, false, false)
	profile.BungieNetMembershipID = currentAccount.BungieNetUser.MembershipID

	loadoutJSON, err := db.SelectLoadout(profile.BungieNetMembershipID, name)
	if err == nil && loadoutJSON == "" {
		raven.CaptureError(errors.New("No loadout matching name"), map[string]string{"name": name})
		outputStr := fmt.Sprintf("Sorry Guardian, you do not have a loadout named %s. "+
			"You need to create a loadout with that name before it can be equipped.", name)
		return outputStr, nil
	} else if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read loadout from the database")
		return "", err
	}

	var peristedLoadout PersistedLoadout
	err = json.NewDecoder(bytes.NewReader([]byte(loadoutJSON))).Decode(&peristedLoadout)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to decode JSON: %s", err.Error())
		return "", err
	}

	loadout := fromPersistedLoadout(peristedLoadout, profile)
	equipLoadout(loadout, profile.Characters[0].CharacterID, profile,
		profile.MembershipType, client)

	outputStr := "All set Guardian, your " + name + " loadout has been restored!"

	return outputStr, nil
}

// GetLoadoutNames will request all of the loadout names from the database and
// return the list of names to the user.
func GetLoadoutNames(accessToken string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(accessToken, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Couldn't load account with the access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return "", errors.New("Couldn't load the current account")
	}

	// Retrieve the existing loadouts from the database
	bnetMembershipID := currentAccount.BungieNetUser.MembershipID
	existing, _ := db.SelectLoadouts(bnetMembershipID)

	if len(existing) == 0 {
		glg.Debug("No loadouts found for user")
		outputStr := "It looks like you don't currently have any saved loadouts." +
			"You can save your current loadout by saying 'Save this loadout for later'"
		return outputStr, nil
	}

	speech := bytes.NewBufferString("Guardian, it looks like you have loadouts saved with these names: ")

	names := strings.Join(existing, ", ")
	speech.WriteString(names)

	return speech.String(), nil
}

// GetCurrentCrucibleRanking will look up the current Glory and Valor rankings for the
// current user and summarize it including the name of the rank and the points
// required to hit the next rank (if they aren't at the top most rank).
func GetCurrentCrucibleRanking(token string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(token, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Couldn't load account with the access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return "", errors.New("Couldn't load the current account")
	}

	membership := currentAccount.DestinyMembership

	progressionResponse := CharacterProgressionResponse{}
	err := client.Execute(NewGetProgressionsRequest(membership.MembershipType,
		membership.MembershipID), &progressionResponse)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Profile response from Bungie!: %s", err.Error())
		return "", errors.New("Failed to read current user's profile: " + err.Error())
	}

	glory := progressionResponse.gloryProgression()
	valor := progressionResponse.valorProgression()

	if glory == nil && valor == nil {
		outputStr := "There was an error trying to find your current Crucible rankings, " +
			" please try again later."
		return outputStr, nil
	}

	buf := bytes.NewBuffer([]byte{})
	if glory != nil {
		var gloryRank string
		if glory.Level < len(pvpRankSteps) {
			gloryRank = pvpRankSteps[glory.Level]
		} else {
			gloryRank = pvpRankSteps[len(pvpRankSteps)-1]
		}
		buf.WriteString(fmt.Sprintf("You've achieved %s glory rank, ", gloryRank))
		if glory.Level == glory.LevelCap {
			buf.WriteString("you have reached max glory, time to reset it and do it again!")
		} else if glory.Level == (glory.LevelCap - 1) {
			buf.WriteString("congratulations on becoming Legend! ")
		} else {
			buf.WriteString(fmt.Sprintf("only %d glory points to the next rank. ",
				glory.NextLevelAt))
		}
	}
	if valor != nil {
		var valorRank string
		if valor.Level < len(pvpRankSteps) {
			valorRank = pvpRankSteps[valor.Level]
		} else {
			valorRank = pvpRankSteps[len(pvpRankSteps)-1]
		}
		buf.WriteString(fmt.Sprintf("Your current valor rank is %s, ", valorRank))
		if valor.Level == valor.LevelCap {
			buf.WriteString("you have reached max valor, time to reset it and do it again!")
		} else if valor.Level == (valor.LevelCap - 1) {
			buf.WriteString("congratulations on becoming Legend! ")
		} else {
			buf.WriteString(fmt.Sprintf("only %d valor points to the next rank. ",
				valor.NextLevelAt))
		}
	}

	buf.WriteString("Goodluck!")

	return buf.String(), nil
}

// GetCurrentGambitRanking will look up the current Infamy (Gambit game mode) ranking for the
// current character including the name and points required to get to the next level.
func GetCurrentGambitRanking(token string) (string, error) {

	client := Clients.Get()
	client.AddAuthValues(token, warmindAPIKey)

	// TODO: check error
	currentAccount, _ := client.GetCurrentAccount()

	if currentAccount == nil {
		raven.CaptureError(errors.New("Couldn't load account with the access token"), nil)
		glg.Error("Failed to load current account with the specified access token!")
		return "", errors.New("Couldn't load the current account")
	}

	membership := currentAccount.DestinyMembership

	progressionResponse := CharacterProgressionResponse{}
	err := client.Execute(NewGetProgressionsRequest(membership.MembershipType,
		membership.MembershipID), &progressionResponse)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read the Profile response from Bungie!: %s", err.Error())
		return "", errors.New("Failed to read current user's profile: " + err.Error())
	}

	infamy := progressionResponse.infamyProgression()

	if infamy == nil {
		outputStr := "There was an error trying to find your current Gambit rankings, " +
			" please try again later."
		return outputStr, nil
	}

	buf := bytes.NewBuffer([]byte{})
	if infamy != nil {
		var infamyRank string
		if infamy.Level < len(gambitRankSteps) {
			infamyRank = gambitRankSteps[infamy.Level]
		} else {
			infamyRank = gambitRankSteps[len(gambitRankSteps)-1]
		}
		buf.WriteString(fmt.Sprintf("You've achieved %s infamy rank, ", infamyRank))
		if infamy.Level == infamy.LevelCap {
			buf.WriteString("you have reached max infamy, time to reset it and do it again!")
		} else if infamy.Level == (infamy.LevelCap - 1) {
			buf.WriteString("congratulations on becoming Legend! ")
		} else {
			buf.WriteString(fmt.Sprintf("only %d infamy points to the next rank. ",
				infamy.NextLevelAt))
		}

		buf.WriteString("Goodluck!")
	}

	return buf.String(), nil
}

// GetOutboundIP gets preferred outbound ip of this machine
// func GetOutboundIP() net.IP {
// 	conn, err := net.Dial("udp", "8.8.8.8:80")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close()

// 	localAddr := conn.LocalAddr().(*net.UDPAddr)

// 	return localAddr.IP
// }

// transferItems is a generic transfer method that will handle a full transfer of a specific item to
// the specified character. This requires a full trip from the source, to the vault, and then to the
// destination character. By providing a nil destCharacter, the items will be transferred to the
// vault and left there. Using a count of -1 will cause all of the items to be transferred.
func transferItems(itemSet []*Item, destCharacter *Character,
	membershipType int, count int, client *Client) int {

	// TODO: This should probably take the transferStatus field into account,
	// if the item is NotTransferrable, don't bother trying.
	var totalCount int
	var wg sync.WaitGroup

	for _, item := range itemSet {

		if item == nil || item.Character == destCharacter {
			// Item is already on the destination character, skipping...
			continue
		}

		numToTransfer := item.Quantity
		if count != -1 {
			numNeeded := count - totalCount
			glg.Debugf("Getting to transfer logic: needed=%d, toTransfer=%d", numNeeded, numToTransfer)
			if numToTransfer > numNeeded {
				numToTransfer = numNeeded
			}
		}
		totalCount += numToTransfer

		wg.Add(1)

		// TODO: There is an issue were we are getting throttling responses from the Bungie
		// servers. There will be an extra delay added here to try and avoid the throttling.
		go func(item *Item, wait *sync.WaitGroup) {

			defer wg.Done()

			glg.Infof("Transferring item: %+v", item)

			// If these items are already in the vault, skip it they will be transferred later
			if item.Character != nil {
				// These requests are all going TO the vault, the FROM the vault request
				// will go later for all of these.
				requestBody := map[string]interface{}{
					"itemReferenceHash": item.ItemHash,
					"stackSize":         numToTransfer,
					"transferToVault":   true,
					"itemId":            item.InstanceID,
					"characterId":       item.Character.CharacterID,
					"membershipType":    membershipType,
				}

				transferClient := Clients.Get()
				transferClient.AddAuthValues(client.AccessToken, client.APIToken)
				response := BaseResponse{}
				transferClient.Execute(NewPostTransferItemRequest(requestBody), &response)
			}

			// TODO: This could possibly be handled more efficiently if we know the items are
			// uniform, meaning they all have the same itemHash values, for example (all motes of
			// light or all strange coins) It is trickier for instances like engrams where each
			// engram type has a different item hash. Now transfer all of these items from the
			// vault to the destination character
			if destCharacter == nil {
				// If the destination is the vault... then we are done already
				return
			}

			vaultToCharRequestBody := map[string]interface{}{
				"itemReferenceHash": item.ItemHash,
				"stackSize":         numToTransfer,
				"transferToVault":   false,
				"itemId":            item.InstanceID,
				"characterId":       destCharacter.CharacterID,
				"membershipType":    membershipType,
			}

			transferClient := Clients.Get()
			transferClient.AddAuthValues(client.AccessToken, client.APIToken)
			response := BaseResponse{}
			transferClient.Execute(NewPostTransferItemRequest(vaultToCharRequestBody), &response)

		}(item, &wg)

		if count != -1 && totalCount >= count {
			break
		}
	}

	wg.Wait()

	return totalCount
}

// equipItems is a generic equip method that will handle a equipping a specific
// item on a specific character.
func equipItems(itemSet []*Item, characterID string,
	membershipType int, client *Client) {

	ids := make([]int64, 0, len(itemSet))

	for _, item := range itemSet {

		if item == nil {
			continue
		}

		if item.TransferStatus == ItemIsEquipped && item.Character.CharacterID == characterID {
			// If this item is already equipped, skip it.
			glg.Debugf("Not equipping item, it is already equipped on the current character: %s",
				item.InstanceID)
			continue
		}

		instanceID, err := strconv.ParseInt(item.InstanceID, 10, 64)
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Not equipping item, the instance ID could not be parsed to an Int:  %s",
				err.Error())
			continue
		}

		glg.Debugf("Equipping item with name : %+v", item)
		ids = append(ids, instanceID)
	}

	glg.Debugf("Equipping items: %+v", ids)

	if len(ids) == 0 {
		glg.Debugf("No items to be equipped, all done.")
		return
	}

	equipRequestBody := map[string]interface{}{
		"itemIds":        ids,
		"characterId":    characterID,
		"membershipType": membershipType,
	}

	// Having a single equip call should avoid the throttling problems.
	response := BaseResponse{}
	client.Execute(NewPostEquipItem(equipRequestBody, true), &response)
}

// TODO: All of these equip/transfer/etc. action should take a single struct with all the
// parameters required to perform the action, as well as probably a *Client reference.

// equipItem will take the specified item and equip it on the provided character
func equipItem(item *Item, character *Character, membershipType int, client *Client) {
	if item == nil {
		glg.Debug("Trying to equip a nil item, ignoring.")
		return
	}

	glg.Debugf("Equipping item(%d, %d)...", item.ItemHash, item.InstanceID)

	equipRequestBody := map[string]interface{}{
		"itemId":         item.InstanceID,
		"characterId":    character.CharacterID,
		"membershipType": membershipType,
	}

	response := BaseResponse{}
	client.Execute(NewPostEquipItem(equipRequestBody, false), &response)
}
