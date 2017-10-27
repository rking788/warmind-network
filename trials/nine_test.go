package trials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseCurrentMap(t *testing.T) {
	data, err := readSample("GenericWeekData.json")

	var currentMap CurrentMap
	err = json.Unmarshal(data, &currentMap)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	if currentMap.Response == nil {
		t.FailNow()
	} else if currentMap.Response.StartDate.String() == "" {
		t.FailNow()
	} else if currentMap.Response.EndDate.String() == "" {
		t.FailNow()
	} else if currentMap.Response.Weapons == nil || len(currentMap.Response.Weapons) != 3 {
		t.FailNow()
	} else if currentMap.Response.Mode == "" || currentMap.Response.MapName == "" {
		t.FailNow()
	}
}

func TestParsePlayerCurrentWeek(t *testing.T) {
	data, err := readSample("PlayerCurrentWeek.json")

	playerCurrentWeek := make([]*PlayerCurrentWeekStats, 0, 1)
	err = json.Unmarshal(data, &playerCurrentWeek)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	if playerCurrentWeek == nil || playerCurrentWeek[0] == nil {
		fmt.Println("Found nil response")
		t.FailNow()
	}

	currentWeek := playerCurrentWeek[0]
	if currentWeek.Map == "" {
		fmt.Println("Empty map string")
		t.FailNow()
	} else if currentWeek.Daily == nil || len(currentWeek.Daily) != 4 {
		fmt.Println("Incorrect daily breakdown")
		t.FailNow()
	} else if currentWeek.Week == "" {
		fmt.Println("Current week number is empty")
		t.FailNow()
	} else if currentWeek.Weapons == nil || len(currentWeek.Weapons) != 3 {
		fmt.Println("Incorrect current week top weapons")
		t.FailNow()
	}
}

func TestParsePlayerSummary(t *testing.T) {
	data, err := readSample("SinglePlayerOverallStats.json")

	var singlePlayerStats PlayerSummaryStats
	err = json.Unmarshal(data, &singlePlayerStats)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	firstResult := singlePlayerStats.Results[0]
	if firstResult.Assists == 0 {
		fmt.Println("Assists is wrong")
		t.FailNow()
	} else if firstResult.Kills == 0 {
		fmt.Println("Kills is wrong")
		t.FailNow()
	} else if firstResult.Deaths == 0 {
		fmt.Println("Deaths is wrong")
		t.FailNow()
	} else if firstResult.Losses == 0 {
		fmt.Println("Losses is wrong")
		t.FailNow()
	} else if firstResult.Matches == 0 {
		fmt.Println("Matches is wrong")
		t.FailNow()
	} else if firstResult.Flawless == 0 {
		fmt.Println("Flawless is wrong")
		t.FailNow()
	} else if firstResult.Badges == nil || len(firstResult.Badges) == 0 {
		fmt.Println("Badges is wrong")
		t.FailNow()
	} else if firstResult.DisplayName == "" {
		fmt.Println("DisplayName is wrong")
		t.FailNow()
	} else if firstResult.MembershipType == 0 {
		t.FailNow()
	} else if firstResult.MembershipID == "" || firstResult.MembershipType == 0 {
		fmt.Println("MembershipID is wrong")
		t.FailNow()
	} else if firstResult.Characters == nil || len(firstResult.Characters) == 0 {
		fmt.Println("Characters is wrong")
		t.FailNow()
	}
}

func TestParseRivalsStats(t *testing.T) {
	data, err := readSample("Rivals.json")

	rivalStats := make([]*RivalsStats, 0, 10)
	err = json.Unmarshal(data, &rivalStats)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	if rivalStats == nil {
		fmt.Println("Found nil unmarshaled slice")
		t.FailNow()
	}
}

func TestParseTeammatesStats(t *testing.T) {
	data, err := readSample("TeammatesMostPlayedWith.json")

	teammatesStats := make([]*TeammatesStats, 0, 10)
	err = json.Unmarshal(data, &teammatesStats)
	if err != nil {
		fmt.Println("Error unmarshaling json: ", err.Error())
		t.FailNow()
	}

	if teammatesStats == nil {
		fmt.Println("Found nil unmarshaled slice")
		t.FailNow()
	}
}

func readSample(name string) ([]byte, error) {
	f, err := os.Open("../local_tools/samples/trials-of-the-nine/" + name)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}
