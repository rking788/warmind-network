package trials

import "time"

// NineClient will be the type that will execute all rquests to the Trials Report API
type NineClient struct {
	BaseURL string
}

// CurrentMap has stats for the current week of Trials of the Nine.
// This data includes the map name and the mode as well as the number
// of players, number of kills, etc.
// http://api.trialsofthenine.com/week/0/
type CurrentMap struct {
	Status   string `json:"Status"`
	Message  string `json:"Message"`
	Response *struct {
		MapName           string                                   `json:"name"`
		Mode              string                                   `json:"mode"`
		PlayerCount       int                                      `json:"players"`
		Matches           int                                      `json:"matches"`
		Week              int                                      `json:"week"`
		Kills             int                                      `json:"kills"`
		StartDate         NineTime                                 `json:"startDate"`
		EndDate           NineTime                                 `json:"endDate"`
		WeaponKills       int                                      `json:"weaponKills"`
		Weapons           map[string][]*nineWeaponStats            `json:"weapons"`
		WeaponsByPlatform map[string]map[string][]*nineWeaponStats `json:"weaponsByPlatform"`
	} `json:"Response"`
}

func (m *CurrentMap) startMonth() string {
	return m.Response.StartDate.Month().String()
}

func (m *CurrentMap) startDay() int {
	return m.Response.StartDate.Day()
}

func (m *CurrentMap) mode() string {
	return m.Response.Mode
}

func (m *CurrentMap) mapName() string {
	return m.Response.MapName
}

type nineWeaponStats struct {
	Bucket     int    `json:"bucket"`
	Headshots  int    `json:"headshots"`
	Matches    int    `json:"matches"`
	Kills      int    `json:"kills"`
	Name       string `json:"name"`
	WeaponType string `json:"weaponType"`
}

// PlayerCurrentWeekStats represents the stats for a particular membership
// for the current week of Trials of the Nine.
// https://api.destinytrialsreport.com/slack/trials/week/4611686018433214701/0
type PlayerCurrentWeekStats struct {
	Map      string `json:"map"`
	Flawless int    `json:"flawless"`
	Week     string `json:"week"`
	Weapons  []*struct {
		SumHeadshots int    `json:"sum_headshots"`
		SumKills     int    `json:"sum_kills"`
		ItemTypeName string `json:"itemTypeDisplayName"`
	} `json:"weapons"`
	Daily []*struct {
		Day     string  `json:"day"`
		Losses  int     `json:"losses"`
		Assists int     `json:"assists"`
		Deaths  int     `json:"deaths"`
		Kills   int     `json:"kills"`
		KD      float64 `json:"kd"`
		Matches int     `json:"matches"`
	} `json:"daily"`
}

func (stats *PlayerCurrentWeekStats) totalMatches() int {
	sum := 0
	for _, day := range stats.Daily {
		sum += day.Matches
	}
	return sum
}

func (stats *PlayerCurrentWeekStats) totalWins() int {
	return stats.totalMatches() - stats.totalLosses()
}

func (stats *PlayerCurrentWeekStats) totalLosses() int {
	sum := 0
	for _, day := range stats.Daily {
		sum += day.Losses
	}
	return sum
}
func (stats *PlayerCurrentWeekStats) totalKills() int {
	sum := 0
	for _, day := range stats.Daily {
		sum += day.Kills
	}

	return sum
}

func (stats *PlayerCurrentWeekStats) totalDeaths() int {
	sum := 0
	for _, day := range stats.Daily {
		sum += day.Deaths
	}
	return sum
}

func (stats *PlayerCurrentWeekStats) combinedKD() float64 {

	return float64(stats.totalKills()) / float64(stats.totalDeaths())
}

// PlayerSummaryStats represents the player's total stats for all
// Trials of the Nine weeks.
// https://api.trialsofthenine.com/player/4611686018433723819/
type PlayerSummaryStats struct {
	Count   int `json:"count"`
	Results []*struct {
		*CharacterStats
		Flawless       int    `json:"Flawless"`
		DisplayName    string `json:"displayName"`
		MembershipType int    `json:"membershipType"`
		MembershipID   string `json:"membershipId"`
		Current        *struct {
			*CharacterStats
			Weapons []*struct{} `json:"weapons"`
		} `json:"current"`
		Characters map[string]*struct {
			*CharacterStats
		} `json:"characters"`
		Activities []*struct {
		} `json:"activities"`
		Badges []string `json:"badges"`
	}
}

// RivalsStats describes the players considered rivals and returned
// by Trials Report
// https://api.trialsofthenine.com/player/4611686018433723819/rivals/
type RivalsStats struct {
	Results []*playerStats
}

// TeammatesStats holds statistics about the teammates a user has
// played with the most.
// https://api.trialsofthenine.com/player/4611686018433723819/teammates/
type TeammatesStats struct {
	Results []*playerStats
}

// CharacterStats will store data about specific players as returned by the Trials Report API
type CharacterStats struct {
	Matches int `json:"matches"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`
	Losses  int `json:"losses"`
	Kills   int `json:"kills"`
}

type playerStats struct {
	DisplayName  string `json:"displayName"`
	MembershipID string `json:"membershipId"`
	*CharacterStats
}

// NineTime is a custom time wrapper that allows the dates
// returned by Trials Report to be parsed with a custom format string.
type NineTime struct {
	time.Time
}

// UnmarshalJSON is responsible for unmarshaling a time.Time in the format provided by
// Trials Report
func (nt *NineTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	// Get rid of the quotes "" around the value.
	// A second option would be to include them
	// in the date format string instead, like so below:
	//   time.Parse(`"`+time.RFC3339Nano+`"`, s)
	s = s[1 : len(s)-1]

	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.999999999Z0700", s)
	}
	nt.Time = t
	return
}
