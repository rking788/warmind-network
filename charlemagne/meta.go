package charlemagne

import (
	"bytes"
	"fmt"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
	"github.com/mikeflynn/go-alexa/skillserver"
)

var (
	// Charlemagne understands the following aggregate modeTypes
	// raid - 4
	// allPvp - 5
	// allPve - 7
	metaActivityTranslation = map[string]string{
		"pvp":                "5",
		"crucible":           "5",
		"trials":             "5",
		"quickplay":          "5",
		"quick play":         "5",
		"competitive":        "5",
		"pve":                "7",
		"strikes":            "7",
		"strike":             "7",
		"story":              "7",
		"faction":            "7",
		"nightfall":          "7",
		"prestige":           "7",
		"prestige nightfall": "7",
		"public event":       "7",
		"adventure":          "7",
		"patrol":             "7",
		"raid":               "4",
		"leviathan":          "4",
		"spire of stars":     "4",
		"eater of worlds":    "4",
	}

	cachedMetaResponses map[string]map[string]*MetaResponse
)

func uniqueModeTypes() []string {
	uniques := make([]string, 0, 10)

	// blank game mode is also possible
	uniques = append(uniques, "")

	for _, modeType := range metaActivityTranslation {
		found := false
		for _, t := range uniques {
			if t == modeType {
				found = true
				break
			}
		}

		if !found {
			uniques = append(uniques, modeType)
		}
	}

	return uniques
}

// MetaResponse contains the fields returned from the Charlemagne API for meta endpoints
type MetaResponse struct {
	*BaseResponse
	*InnerMetaResponse `json:"response"`
}

// InnerMetaResponse is a wrapper around the inner response JSON tag. This is where
// the majority of the data is stored.
type InnerMetaResponse struct {
	WeekNumber        int            `json:"weekNumber"`
	ActivityHash      string         `json:"activityHash"`
	MembershipTypes   string         `json:"membershipTypes"` // Comma separated
	ModeTypes         string         `json:"modeTypes"`       // Comma separated
	TopKineticWeapons []*WeaponStats `json:"topKineticWeapons"`
	TopEnergyWeapons  []*WeaponStats `json:"topEnergyWeapons"`
	TopPowerWeapons   []*WeaponStats `json:"topPowerWeapons"`
}

// WeaponStats summarizes statistics about a particular weapon in the meta responses
type WeaponStats struct {
	WeaponID   int    `json:"weaponId"`
	WeaponName string `json:"weaponName"`
	TotalKills int    `json:"totalKills"`
	WeaponType string `json:"weaponType"`
}

// FindCurrentMeta will return information about commonly used weapons via the Charlemagne API
func FindCurrentMeta(platform, requestedActivity string) (*skillserver.EchoResponse, error) {

	response := skillserver.NewEchoResponse()

	// TODO: activity hash is currently unused. may be useful to be able to request meta
	// for a current activity
	//activityHash := ""
	gameModes := []string{}
	translatedGameMode, ok := metaActivityTranslation[requestedActivity]
	if ok {
		// At some ponit this could be a list of activities but for now let's leave it as one
		gameModes = append(gameModes, translatedGameMode)
		glg.Infof("Found translated game mode of: %s", translatedGameMode)
	}
	translatedPlatform := platformNameToMapKey[platform]

	glg.Warnf("Translated game mode for lookup in cache: %s", translatedGameMode)
	meta := cachedMetaResponses[translatedPlatform][translatedGameMode]
	if meta == nil {
		err := fmt.Errorf("Trying to return meta for mode=%s and platform=%s but not in cache", translatedGameMode, translatedPlatform)
		raven.CaptureError(err, nil, nil)
		return nil, err
	}

	speechBuffer := bytes.NewBuffer([]byte{})
	speechBuffer.Write([]byte("The most commonly used weapons according to Charlemagne "))
	if requestedActivity != "" {
		speechBuffer.WriteString(fmt.Sprintf("for %s ", requestedActivity))
	}
	if platform != "" {
		speechBuffer.WriteString(fmt.Sprintf("on %s ", platform))
	}

	speechBuffer.WriteString(fmt.Sprintf("include %s and %s for kinetic weapons,",
		meta.TopKineticWeapons[0].WeaponName, meta.TopKineticWeapons[1].WeaponName))
	speechBuffer.WriteString(fmt.Sprintf("%s and %s for energy weapons,",
		meta.TopEnergyWeapons[0].WeaponName, meta.TopEnergyWeapons[1].WeaponName))
	speechBuffer.WriteString(fmt.Sprintf("and %s and %s for power weapons.",
		meta.TopPowerWeapons[0].WeaponName, meta.TopPowerWeapons[1].WeaponName))

	response.OutputSpeech(speechBuffer.String())

	return response, nil
}
