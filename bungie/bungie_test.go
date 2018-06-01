package bungie

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/kpango/glg"

	"github.com/rking788/warmind-network/db"
)

func setup() {
	glg.Get().SetLevelMode(glg.DEBG, glg.NONE)
	glg.Get().SetLevelMode(glg.INFO, glg.NONE)
	glg.Get().SetLevelMode(glg.WARN, glg.NONE)

	db.InitEnv(os.Getenv("DATABASE_URL"))
	InitEnv("", "")
}

// NOTE: Never run this while using the bungie.net URLs in bungie/constants.go
// those should be changed to a localhost webserver that returns static results.
// func BenchmarkSomething(b *testing.B) {

// 	profileResponse, err := getCurrentProfileResponse()
// 	if err != nil {
// 		b.Fail()
// 		return
// 	}
// 	_ = fixupProfileFromProfileResponse(profileResponse)

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		//CountItem("strange coins", "aaabbbccc")
// 	}
// }

func BenchmarkFilteringSingleFilter(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items._FilterItems(itemTierTypeFilter, ExoticTier)
	}
}
func BenchmarkFilteringMultipleFilters(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.
			_FilterItems(itemClassTypeFilter, WARLOCK).
			_FilterItems(itemNotTierTypeFilter, ExoticTier).
			_FilterItems(itemRequiredLevelFilter, 25)
	}
}

func BenchmarkFilteringMultipleFiltersAtOnce(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.
			_FilterItemsMultiple(createItemClassTypeFilter(WARLOCK),
				createItemNotTierTypeFilter(ExoticTier),
				createItemRequiredLevelFilter(25))
	}
}

func BenchmarkFilteringMultipleFiltersAtOnceBubble(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.
			FilterItemsMultipleBubble(createItemClassTypeFilter(WARLOCK),
				createItemNotTierTypeFilter(ExoticTier),
				createItemRequiredLevelFilter(25))
	}
}

func BenchmarkFilteringPassthrough(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.
			FilterBaseline(createItemClassTypeFilter(WARLOCK),
				createItemNotTierTypeFilter(ExoticTier),
				createItemRequiredLevelFilter(25))
	}
}

func (items ItemList) FilterBaseline(filters ...ItemFilter) ItemList {

	for i, item := range items {
		if item.InstanceID == "0" {
			fmt.Println("Breaking")
			break
		}

		tempItem := items[i]
		items[i] = items[len(items)-i-1]
		items[len(items)-i-1] = tempItem
	}

	return make(ItemList, 10)
}
func TestFilteringSingleFilterBubble(t *testing.T) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems

	resultNormal := items._FilterItems(itemTierTypeFilter, ExoticTier)
	resultBubble := items.FilterItemsBubble(itemTierTypeFilter, ExoticTier)

	fmt.Printf("Found normal(%d), bubble(%d)", len(resultNormal), len(resultBubble))

	if len(resultNormal) != len(resultBubble) {
		t.Errorf("Incorrect item count normal(%d) bubble(%d)", len(resultNormal), len(resultBubble))
		t.FailNow()
	}
}

func TestFilteringMultipleFilter(t *testing.T) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems

	resultNormal := items.
		_FilterItems(itemClassTypeFilter, WARLOCK).
		_FilterItems(itemNotTierTypeFilter, ExoticTier).
		_FilterItems(itemRequiredLevelFilter, 25)
	resultBubble := items.
		FilterItemsBubble(itemClassTypeFilter, WARLOCK).
		FilterItemsBubble(itemNotTierTypeFilter, ExoticTier).
		FilterItemsBubble(itemRequiredLevelFilter, 25)
	resultSingleFilter := items.
		FilterItemsMultipleBubble(createItemClassTypeFilter(WARLOCK),
			createItemNotTierTypeFilter(ExoticTier),
			createItemRequiredLevelFilter(25))

	fmt.Printf("Found normal(%d), bubble(%d), single(%d)\n", len(resultNormal), len(resultBubble), len(resultSingleFilter))

	if len(resultNormal) != len(resultBubble) {
		t.Errorf("Incorrect item count normal(%d) bubble(%d)", len(resultNormal), len(resultBubble))
		t.FailNow()
	}
	if len(resultNormal) != len(resultSingleFilter) {
		t.Errorf("Incorrect item count normal(%d) singleFilter(%d)", len(resultNormal),
			len(resultSingleFilter))
		t.FailNow()
	}

	foundCount := len(resultNormal)
	for _, outer := range resultNormal {
		found := false

		for _, inner := range resultSingleFilter {
			if outer.InstanceID == inner.InstanceID {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Item(%d) with instanceID(%s) not found in single bubble results",
				outer.ItemHash, outer.InstanceID)
			t.FailNow()
		}
	}

	if foundCount != len(resultSingleFilter) {
		t.Errorf("All elements from the normal filter method were not found in the " +
			"single bubble filter results")
		t.FailNow()
	}
}

func BenchmarkFilteringSingleFilterBubble(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.FilterItemsBubble(itemTierTypeFilter, ExoticTier)
	}
}
func BenchmarkFilteringMultipleFiltersBubble(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)

	items := profile.AllItems
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = items.
			FilterItemsBubble(itemClassTypeFilter, WARLOCK).
			FilterItemsBubble(itemNotTierTypeFilter, ExoticTier).
			FilterItemsBubble(itemRequiredLevelFilter, 25)
	}
}

func BenchmarkMaxLight(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, true)
	testDestinationID := profile.Characters[0].CharacterID

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findMaxLightLoadout(profile, testDestinationID)
	}
}
func BenchmarkFindRandomLoadoutWeaponsOnly(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)
	testDestinationID := profile.Characters[0].CharacterID

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findRandomLoadout(profile, testDestinationID, false)
	}
}

func BenchmarkFindRandomLoadoutAll(b *testing.B) {

	setup()
	profileResponse, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
		return
	}
	profile := fixupProfileFromProfileResponse(profileResponse, false)
	testDestinationID := profile.Characters[0].CharacterID

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findRandomLoadout(profile, testDestinationID, true)
	}
}

func BenchmarkGroupAndSort(b *testing.B) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
	}
	profile := fixupProfileFromProfileResponse(response, false)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile := groupAndSortGear(profile.AllItems)
		if profile == nil {
			b.FailNow()
		}
	}
}

func BenchmarkBestItemForBucket(b *testing.B) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
	}
	profile := fixupProfileFromProfileResponse(response, false)
	grouped := groupAndSortGear(profile.AllItems)
	largestBucket := Kinetic
	largestBucketSize := len(grouped[Kinetic])
	for bkt, list := range grouped {
		if len(list) > largestBucketSize {
			largestBucket = bkt
			largestBucketSize = len(list)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := findBestItemForBucket(largestBucket, grouped[largestBucket], profile.Characters[0].CharacterID)
		if item == nil {
			b.FailNow()
		}
	}
}

func BenchmarkFixupProfileFromProfileResponse(b *testing.B) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		b.FailNow()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile := fixupProfileFromProfileResponse(response, false)
		if profile == nil {
			b.FailNow()
		}
	}
}

func TestParseCurrentMembershipsResponse(t *testing.T) {
	setup()
	data, err := readSample("GetMembershipsForCurrentUser.json")
	if err != nil {
		fmt.Println("Error reading sample file: ", err.Error())
		t.FailNow()
	}

	var response CurrentUserMembershipsResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	if response.Response.BungieNetUser == nil {
		t.FailNow()
	}

	if response.Response.DestinyMemberships == nil {
		t.FailNow()
	}
	if len(response.Response.DestinyMemberships) != 2 {
		t.FailNow()
	}
	for _, membership := range response.Response.DestinyMemberships {
		if membership.DisplayName == "" || membership.MembershipID == "" || membership.MembershipType <= 0 {
			t.FailNow()
		}
	}
}

func TestParseGetProfileResponse(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}

	if response.Response.Profile == nil || response.Response.ProfileCurrencies == nil ||
		response.Response.ProfileInventory == nil || response.Response.CharacterEquipment == nil ||
		response.Response.CharacterInventories == nil || response.Response.Characters == nil {
		fmt.Println("One of the expected entries is nil!")
		t.FailNow()
	}

	if len(response.Response.Characters.Data) != 3 {
		fmt.Println("Character count incorrect")
		t.FailNow()
	}

	if len(response.Response.ProfileCurrencies.Data.Items) != 3 {
		fmt.Println("Currency count incorrect")
		t.FailNow()
	}

	if len(response.Response.CharacterEquipment.Data) == 0 || len(response.Response.CharacterInventories.Data) == 0 {
		fmt.Println("Incorrect number of character equipment and character inventory items.")
		t.FailNow()
	}

	for _, char := range response.Response.CharacterEquipment.Data {
		for _, item := range char.Items {
			if item.InstanceID == "" {
				fmt.Println("Found a character equiment item without an instance ID")
				t.FailNow()
			}
		}
	}

	if response.Response.ProfileCurrencies.Data.Items[0].InstanceID != "" {
		fmt.Println("Found a profile currency entry with an instance ID")
		t.FailNow()
	}
}

func TestFixupProfileFromProfileResponse(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}
}

func TestFixupProfileFromProfileResponseOnlyInstanced(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}
}

func TestFixupProfileFromProfileResponseMissingProfile(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}
	response.Response.Profile = nil

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}

	if profile.MembershipID != "" {
		t.FailNow()
	}
	if profile.MembershipType != 0 {
		t.FailNow()
	}
}

func TestFixupProfileFromProfileResponseMissingProfileInventory(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}
	response.Response.ProfileInventory = nil

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}
}

func TestFixupProfileFromProfileResponseMissingCharacters(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}
	response.Response.Characters = nil

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}

	if profile.Characters != nil {
		t.FailNow()
	}

	for _, item := range profile.AllItems {
		if item.Character != nil {
			t.FailNow()
		}
	}
}

func TestFixupProfileFromProfileResponseMissingCharacterEquipment(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}
	response.Response.CharacterEquipment = nil

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}

	for _, item := range profile.AllItems {
		if item.ItemInstance != nil && item.IsEquipped == true {
			t.FailNow()
		}
	}
}

func TestFixupProfileFromProfileResponseMissingCharacterInventories(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}
	response.Response.CharacterInventories = nil

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}
}

func TestGroupAndSort(t *testing.T) {
	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.Fatal("Failed to get current profile for test")
	}

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.Fatal("Failed to fixup profile from response")
	}

	groupedAndSorted := groupAndSortGear(profile.AllItems)
	for bkt, items := range groupedAndSorted {
		targetHash := BucketHashLookup[bkt]
		lastPower := math.MaxInt64

		for _, item := range items {
			if item.BucketHash != targetHash {
				t.Fatalf("Item did not have the correct bucket hash: Required=%d, Found=%d", targetHash, item.BucketHash)
			}

			if lastPower < item.Power() {
				t.Fatalf("Found new Power higher than a previous value: Previous=%d, Current=%d", lastPower, item.Power())
			}

			lastPower = item.Power()
		}
	}
}

func TestRandomLoadoutFromProfile(t *testing.T) {

	setup()
	response, err := getCurrentProfileResponse()
	if err != nil {
		t.FailNow()
	}

	profile := fixupProfileFromProfileResponse(response, false)
	if profile == nil {
		t.FailNow()
	}

	startingLoadout := profile.Loadouts[profile.Characters[0].CharacterID]
	startingInstanceIDs := make([]string, 0, len(startingLoadout))
	for i := Kinetic; i < Artifact; i++ {
		startingInstanceIDs = append(startingInstanceIDs, startingLoadout[i].InstanceID)
	}

	loadout := findRandomLoadout(profile, profile.Characters[0].CharacterID, true)

	for equipmentBucket, item := range loadout {
		if item == nil {
			t.FailNow()
		}

		if _, ok := BucketHashLookup[equipmentBucket]; !ok {
			t.FailNow()
		}
	}

	endingInstanceIDs := make([]string, 0, len(loadout))
	for i := Kinetic; i < Artifact; i++ {
		endingInstanceIDs = append(endingInstanceIDs, loadout[i].InstanceID)
	}

	allTheSame := true
	for index, instanceID := range startingInstanceIDs {
		allTheSame = allTheSame && (instanceID == endingInstanceIDs[index])
	}

	if allTheSame {
		fmt.Println("Calculated a random loadout that was exactly equal to the starting loadout")
		t.FailNow()
	}
}

func TestParseCharacterProgressionsResponse(t *testing.T) {

	data, err := readSample("GetProgressions.json")
	if err != nil {
		t.Fatalf("Error reading in sample progressions file: %s\n", err.Error())
	}

	var response CharacterProgressionResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("Error unmarshaling character progressions: %s\n", err.Error())
	}

	if len(response.Response.CharacterProgressions.Data) == 0 {
		t.Fatal("Empty character progression data from sample response")
	}

	for charID, charProgress := range response.Response.CharacterProgressions.Data {
		if charID == "" {
			t.Fatal("Found empty character ID when iterating through character progressions")
		}

		if len(charProgress.Progressions) == 0 {
			t.Fatalf("Found empty Destiny progression when iterating through character(%s)\n", charID)
		}

		if len(charProgress.Factions) == 0 {
			t.Fatalf("Found empty faction progression when iterating through character(%s)", charID)
		}

		if len(charProgress.Milestones) == 0 {
			t.Fatalf("Found empty milestones progression when iterating through character(%s)", charID)
		}
	}
}

func TestProgressionAccessors(t *testing.T) {

	resp, err := getProgressions()
	if err != nil {
		t.Fatalf("Error reading in sample progressions file: %s\n", err.Error())
	}

	if len(resp.Response.CharacterProgressions.Data) == 0 {
		t.Fatal("Empty character progression data from sample response\n")
	}

	valorBlank := resp.valorProgressionForChar("<blank>")
	if valorBlank != nil {
		t.Fatal("Valor accessor returned a non-nil progression for bogus character ID\n")
	}
	gloryBlank := resp.gloryProgressionForChar("<blank>")
	if gloryBlank != nil {
		t.Fatal("Glory accessor returned a non-nil progression for bogus character ID\n")
	}

	for charID, charProgress := range resp.Response.CharacterProgressions.Data {
		valor := resp.valorProgressionForChar(charID)
		if valor == nil {
			t.Fatal("Found nil valor from accessor with a valid character ID\n")
		}
		if valor != charProgress.Progressions[valorHash] || fmt.Sprintf("%d", valor.ProgressionHash) != valorHash {
			t.Fatal("Valor accessor is potentially returning the progression for " +
				"the wrong character\n")
		}
		fmt.Printf("Valor: %+v\n", valor)

		glory := resp.gloryProgressionForChar(charID)
		if glory == nil {
			t.Fatal("Found nil glory from accessor with a valid character ID\n")
		}
		if glory != charProgress.Progressions[gloryHash] || fmt.Sprintf("%d", glory.ProgressionHash) != gloryHash {
			t.Fatal("Glory accessor is potentially returning glory for the wrong character\n")
		}
		fmt.Printf("Glory: %+v\n", glory)
	}
}

func getCurrentProfileResponse() (*GetProfileResponse, error) {
	data, err := readSample("GetProfile.json")
	if err != nil {
		fmt.Println("Error reading sample file: ", err.Error())
		return nil, err
	}

	var response GetProfileResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		return nil, err
	}

	return &response, nil
}

func getProgressions() (*CharacterProgressionResponse, error) {
	data, err := readSample("GetProgressions.json")
	if err != nil {
		fmt.Println("Error reading sample file: ", err.Error())
		return nil, err
	}

	var response CharacterProgressionResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		return nil, err
	}

	return &response, nil
}

func readSample(name string) ([]byte, error) {
	f, err := os.Open("../test_data/bungie/" + name)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// NOTE: This should never really be run normally. Really just for testing a
// full get profile request chain
// func TestFullGetItemCountTest(t *testing.T) {

// 	setup()

// 	testAccessToken := "<TEST_ACCESS_TOKEN>"
// 	client := Clients.Get()

// 	client.AddAuthValues(testAccessToken, "<TEST_API_KEY>")

// 	// Load all items on all characters
// 	_, _ = GetProfileForCurrentUser(client, false)
// }
