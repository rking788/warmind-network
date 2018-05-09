package charlemagne

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseActivityResponse(t *testing.T) {

	resp, err := getActivityResponse("playerActivity.json")
	if err != nil {
		t.Fatal("Failed to read player activity JSON mock response")
	}

	// Assert that activity fields look correct
	if resp.AverageConcurrentPlayers == 0 {
		t.Fatal("Failed to parse average concurrent players")
	}

	if len(resp.ActivityByMode) != 8 {
		t.Fatal("Unexpected number of activities by mode parsed")
	}

	if len(resp.ActivityByRaid) != 3 {
		t.Fatal("Unexpected number of activities by Raid parsed")
	}

	if len(resp.ActivityByPlatform) != 3 {
		t.Fatal("Unexpected number of activies by platform")
	}

	validateActivityByModeType(resp.ActivityByMode, t)
	validateActivityByRaid(resp.ActivityByRaid, t)

	for platform, summary := range resp.ActivityByPlatform {
		if platform == "" {
			t.Fatal("Found empty platform identifier")
		}
		validateActivityByModeType(summary.ActivityByMode, t)
		validateActivityByRaid(summary.ActivityByRaid, t)
	}
}

func validateActivityByModeType(stats map[string]*ActivitySummary, t *testing.T) {
	expectedModes := []string{"competitive", "iron-banner", "nightfall", "patrol", "quickplay", "raid", "story", "strikes"}
	for _, mode := range expectedModes {
		if val, ok := stats[mode]; ok {
			if val.Mode == 0 {
				t.Fatal("Found a missing/zero mode type in activityByModeType")
			}
			if val.PercentagePlayed == 0.0 {
				t.Fatal("Found a missing/zero percentagePlayer in activityByModeType")
			}
			if val.RawScore == 0 {
				t.Fatal("Found a missing/zero rawScore in actviityByModeType")
			}
		} else {
			t.Fatalf("Did not find %s in the activities by mode map", mode)
		}
	}
}

func validateActivityByRaid(summary map[string]*ActivitySummary, t *testing.T) {
	expectedRaidHashes := []string{
		"1685065161",
		"2693136600",
		"3089205900",
	}

	for _, hash := range expectedRaidHashes {
		if val, ok := summary[hash]; ok {
			if val.ActivityHash == "" || val.ActivityHash != hash {
				t.Fatalf("Found unexpected activity hash in activityByRaid: %s", val.ActivityHash)
			}
			if val.Mode != 0 {
				t.Fatal("Unexpectedly found a non-zero mode value in activityByRaid")
			}
			if val.PercentagePlayed == 0 {
				t.Fatal("Found zero percentagePlayed in activityByRaid")
			}
			if val.RawScore == 0 {
				t.Fatal("Found zero rawScore in activityByRaid")
			}
		} else {
			t.Fatalf("Did not find any activity for the following raid hash: %s", hash)
		}
	}
}

func TestSortedActivityByMode(t *testing.T) {

	response, err := getActivityResponse("playerActivity.json")
	if err != nil {
		t.Fatal("Couldn't load player activity JSON mock")
	}

	sorted := sortPlayerActivityModes(response.ActivityByMode)
	prev := sorted[0]

	for _, activity := range sorted {
		if activity.PercentagePlayed > prev.PercentagePlayed {
			t.Fatal("Not actually sorted, current is greater than the previous percentage played")
		}

		if _, ok := modeTypeToName[activity.Mode]; !ok {
			t.Fatalf("Could not find %d mode value in mode type lookup table", activity.Mode)
		}

		prev = activity
	}
}

func getActivityResponse(name string) (*ActivityResponse, error) {
	data, err := readTestResponse(name)
	if err != nil {
		return nil, err
	}

	response := &ActivityResponse{}
	err = json.Unmarshal(data, response)

	return response, err
}

func readTestResponse(filename string) ([]byte, error) {
	f, err := os.Open("../test_data/charlemagne/" + filename)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}
