package charlemagne

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseMetaResponse(t *testing.T) {
	resp, err := getMetaResponse("meta.json")
	if err != nil {
		t.Fatal("Failed to read player activity JSON mock response")
	}

	if resp == nil {
		t.Fatal("Failed to get a mock response for meta object")
	}

	if resp.WeekNumber == 0 {
		t.Fatal("Found 0 value for the week number")
	}

	validateWeaponActivity(resp.TopKineticWeapons, t)
	validateWeaponActivity(resp.TopEnergyWeapons, t)
	validateWeaponActivity(resp.TopPowerWeapons, t)
}

func TestParseMetaResponseWithActivity(t *testing.T) {
	resp, err := getMetaResponse("meta-with-activity.json")
	if err != nil {
		t.Fatal("Failed to read player activity JSON mock response")
	}

	if resp == nil {
		t.Fatal("Failed to get a mock response for meta object")
	}

	if resp.WeekNumber == 0 {
		t.Fatal("Found 0 value for the week number")
	}

	// This will also have an activity hash field
	if resp.ActivityHash == "" {
		t.Fatal("Found an empty activity hash when parsing meta with activity")
	}

	validateWeaponActivity(resp.TopKineticWeapons, t)
	validateWeaponActivity(resp.TopEnergyWeapons, t)
	validateWeaponActivity(resp.TopPowerWeapons, t)
}

func TestParseMetaResponseWithModes(t *testing.T) {
	resp, err := getMetaResponse("meta-with-modes.json")
	if err != nil {
		t.Fatal("Failed to read player activity JSON mock response")
	}

	if resp == nil {
		t.Fatal("Failed to get a mock response for meta object")
	}

	if resp.WeekNumber == 0 {
		t.Fatal("Found 0 value for the week number")
	}

	// This will also have a list of game mode types
	if resp.ModeTypes == "" {
		t.Fatal("Found an empty mode types string when parsing meta with activity")
	}

	types := strings.Split(resp.ModeTypes, ",")
	if len(types) <= 1 {
		t.Fatal("Did not find multiple game modes in meta response, " +
			"maybe not a comma separated list anymore")
	}

	validateWeaponActivity(resp.TopKineticWeapons, t)
	validateWeaponActivity(resp.TopEnergyWeapons, t)
	validateWeaponActivity(resp.TopPowerWeapons, t)
}

func validateWeaponActivity(stats []*WeaponStats, t *testing.T) {

	for _, stat := range stats {
		if stat.WeaponID == 0 {
			t.Fatal("Found a zero weapon identifier")
		}
		if stat.WeaponName == "" {
			t.Fatal("Empty weapon name")
		}
		if stat.TotalKills == 0 {
			t.Fatal("Found a zero total kill count for a weapon")
		}
		if stat.WeaponType == "" {

		}
	}
}

func getMetaResponse(name string) (*MetaResponse, error) {
	data, err := readTestResponse(name)
	if err != nil {
		return nil, err
	}

	response := &MetaResponse{}
	err = json.Unmarshal(data, response)

	return response, err
}
