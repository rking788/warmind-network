package trials

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/kpango/glg"
	"github.com/rking788/go-alexa/skillserver"
	"github.com/rking788/warmind-network/bungie"
)

const (
	// RequestOrigin will be used in the Origin header when making requests to Trials Report
	RequestOrigin = "https://warmind-network.herokuapp.com"
)

var bungieAPIKey string
var warmindAPIKey string
var client *NineClient

// InitEnv provides a package level initialization point for any work that is environment specific
func InitEnv(apiKey, warmindKey, baseURL string) {
	bungieAPIKey = apiKey
	warmindAPIKey = warmindKey
	if baseURL == "" {
		client = &NineClient{
			BaseURL: NineBaseURL,
		}
	} else {
		client = &NineClient{
			BaseURL: baseURL,
		}
	}
}

// GetCurrentMap will make a request to the Trials Report API endpoint and
// return an Alexa response describing the current map.
func GetCurrentMap() (*skillserver.EchoResponse, error) {

	response := skillserver.NewEchoResponse()

	resp := CurrentMap{}
	err := client.Execute(NineCurrentWeekStatsPath, &resp)

	if err != nil {
		glg.Errorf("Failed to read the current map from Trials Report!: %s", err.Error())
		return nil, err
	}

	response.OutputSpeech(fmt.Sprintf("According to Trials Report, the current Trials of "+
		"the Nine map beginning %s %d is %s, goodluck Guardian.", resp.Response.StartDate.Month().String(), resp.Response.StartDate.Day(), resp.Response.MapName))

	return response, nil
}

// GetCurrentWeek is responsible for requesting the players stats from the current
// week from Trials Report.
func GetCurrentWeek(token, appName string) (*skillserver.EchoResponse, error) {
	response := skillserver.NewEchoResponse()

	membershipID, err := findMembershipID(token, appName)
	if err != nil {
		return nil, err
	}

	fullEndpoint := fmt.Sprintf(NinePlayerCurrentWeekStatsPathFmt, membershipID)
	currentWeek := PlayerCurrentWeekStats{}
	client.Execute(fullEndpoint, &currentWeek)

	if currentWeek.totalMatches() != 0 {
		response.OutputSpeech(fmt.Sprintf("So far you have played %d matches with %d wins, "+
			"%d losses and a combined KD of %.1f, according to Trials Report",
			currentWeek.totalMatches(), currentWeek.totalWins(), currentWeek.totalLosses(), currentWeek.combinedKD()))
	} else {
		response.OutputSpeech("You have not yet played any Trials of the Nine matches this week guardian.")
	}

	return response, nil
}

// GetPopularWeaponTypes will hit the Trials Report endpoint to load info about which weapon
// types are getting the most kills
func GetPopularWeaponTypes() (*skillserver.EchoResponse, error) {

	response := skillserver.NewEchoResponse()

	currentMapInfo := CurrentMap{}
	err := client.Execute(NineCurrentWeekStatsPath, &currentMapInfo)
	if err != nil {
		return nil, err
	}

	kinetic := currentMapInfo.Response.Weapons[string(bungie.BucketHashLookup[bungie.Kinetic])]
	energy := currentMapInfo.Response.Weapons[string(bungie.BucketHashLookup[bungie.Energy])]

	output := fmt.Sprintf("For kinetics it looks like %ss and %ss are the most popular "+
		"this week. %ss and %ss seem to be the most popular energy weapons acoording to "+
		"Trials Report. Goodluck Guardian!", kinetic[0].WeaponType, kinetic[1].WeaponType,
		energy[0].WeaponType, energy[1].WeaponType)

	response.OutputSpeech(output)
	return response, nil
}

// GetWeaponUsagePercentages will return a response describing the top 3 used weapons
// by all players for the current week.
func GetWeaponUsagePercentages() (*skillserver.EchoResponse, error) {
	response := skillserver.NewEchoResponse()

	currentWeekStats := CurrentMap{}
	err := client.Execute(NineCurrentWeekStatsPath, &currentWeekStats)
	if err != nil {
		return nil, err
	}

	kinetics := currentWeekStats.Response.Weapons[string(bungie.BucketHashLookup[bungie.Kinetic])]
	energies := currentWeekStats.Response.Weapons[string(bungie.BucketHashLookup[bungie.Energy])]

	combinedWeapons := make([]*nineWeaponStats, 0, len(kinetics)+len(energies))
	combinedWeapons = append(combinedWeapons, kinetics...)
	combinedWeapons = append(combinedWeapons, energies...)

	sort.Slice(combinedWeapons, func(i, j int) bool {
		return combinedWeapons[i].Matches > combinedWeapons[j].Matches
	})

	buffer := bytes.NewBufferString("According to Trials Report, the top weapons used " +
		"in trials this week are: ")
	// TODO: Maybe it would be good to have the user specify the number of top weapons
	// they want returned.
	for i := 0; i < TopWeaponUsageLimit; i++ {
		usagePercent := float64(combinedWeapons[i].Matches) / float64(currentWeekStats.Response.Matches)
		buffer.WriteString(fmt.Sprintf("%s with %.1f%% of matches, ",
			combinedWeapons[i].Name, usagePercent))
	}

	response.OutputSpeech(buffer.String())
	return response, nil
}

// GetPersonalTopWeapons will return a summary of the top weapons used by the
// linked player/account.
func GetPersonalTopWeapons(token, appName string) (*skillserver.EchoResponse, error) {
	response := skillserver.NewEchoResponse()

	membershipID, err := findMembershipID(token, appName)
	if err != nil {
		glg.Errorf("Error loading membership ID for linked account: %s", err.Error())
		return nil, err
	}

	fullEndpoint := fmt.Sprintf(NinePlayerCurrentWeekStatsPathFmt, membershipID)
	currentWeekStats := PlayerCurrentWeekStats{}
	client.Execute(fullEndpoint, &currentWeekStats)

	buffer := bytes.NewBufferString("According to Trials Report, your top weapons by kills are: ")
	for index, weaponStat := range currentWeekStats.Weapons {

		if index >= TopWeaponUsageLimit || index >= len(currentWeekStats.Weapons) {
			break
		}

		buffer.WriteString(fmt.Sprintf("%s, ", weaponStat.ItemTypeName))
	}

	response.OutputSpeech(buffer.String())

	return response, nil
}

// Execute is a generic method for a NineClient to make HTTP requests to the Trials Report
// API. This method will fully configure the HTTP request as needed including request body.
func (client *NineClient) Execute(endpoint string, response interface{}) error {

	var req *http.Request

	req, _ = http.NewRequest("GET", client.BaseURL+endpoint, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", RequestOrigin)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glg.Errorf("Failed to read the current week stats response from Trials Report!: %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		glg.Errorf("Error parsing trials report response: %s", err.Error())
		return err
	}

	return nil
}

// findMembershipID is a helper function for loading the membership ID from the currently
// linked account, this eventually should take platform into account.
func findMembershipID(token, appName string) (string, error) {

	client := bungie.Clients.Get()
	if appName == "guardian-helper" {
		client.AddAuthValues(token, bungieAPIKey)
	} else {
		client.AddAuthValues(token, warmindAPIKey)
	}
	currentAccount, err := client.GetCurrentAccount()
	if err != nil {
		glg.Errorf("Error loading current account info from Bungie.net: %s", err.Error())
		return "", err
	} else if currentAccount == nil || currentAccount.DestinyMembership == nil {
		return "", errors.New("No linked Destiny account found on Bungie.net")
	}

	return currentAccount.DestinyMembership.MembershipID, nil
}
