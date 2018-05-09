package charlemagne

import (
	"bytes"
	"fmt"

	"github.com/rking788/go-alexa/skillserver"
)

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

	activityHash := ""
	gameModes := []string{}
	membershipType := 0

	meta, err := client.GetCurrentMeta(activityHash, gameModes, membershipType)
	if err != nil {
		return nil, err
	}

	speechBuffer := bytes.NewBuffer([]byte{})
	speechBuffer.Write([]byte("The most commonly used weapons "))
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
