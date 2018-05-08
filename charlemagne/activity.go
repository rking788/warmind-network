package charlemagne

import (
	"fmt"
	"sort"

	"github.com/rking788/go-alexa/skillserver"
)

// ActivityResponse contains all the stats returned by the Charlemagne API
// from the activity endpoints
type ActivityResponse struct {
	*BaseResponse
	*InnerActivityResponse `json:"response"`
}

// InnerActivityResponse is a wrapper around the data inside the inner response
// JSON tag. This is where the majority of the (meaningful) data is stored.
type InnerActivityResponse struct {
	AverageConcurrentPlayers int                         `json:"averageConcurrentPlayers"`
	ActivityByMode           map[string]*ActivitySummary `json:"activityByModeType"`
	ActivityByRaid           map[string]*ActivitySummary `json:"activityByRaid"`
	ActivityByPlatform       map[string]*PlatformSummary `json:"activityByPlatform"`
}

// ActivitySummary contains data that summarizes the activity stats. It seems
// (from the sample responses) that either activity hash or mode will be provided but not both
type ActivitySummary struct {
	ActivityHash     string  `json:"activityHash"`
	Mode             int     `json:"mode"`
	PercentagePlayed float64 `json:"percentagePlayed"`
	RawScore         int     `json:"rawScore"`
}

// PlatformSummary contains the activity data for the different platfomrs. These
// summaries are keyed by the platform ID string.
type PlatformSummary struct {
	MembershipType int                         `json:"membershipType"`
	DisplayName    string                      `json:"displayName"`
	RawScore       int                         `json:"rawScore"`
	Percent        float64                     `json:"percent"`
	ActivityByMode map[string]*ActivitySummary `json:"activityByModeType"`
	ActivityByRaid map[string]*ActivitySummary `json:"activityByRaid"`
}

var (
	modeTypeToName = map[int]string{
		2:  "story",
		3:  "strike",
		4:  "raid",
		6:  "patrol",
		8:  "quickplay",
		9:  "competitive",
		16: "nightfall",
		19: "iron-banner",
	}
	platformNameToMapKey = map[string]string{
		"playstation":  "2",
		"play station": "2",
		"psn":          "2",
		"xbox":         "1",
		"microsoft":    "1",
		"pc":           "4",
	}
)

// FindMostPopularActivities will generate an EchoResponse that describes the most popular
// activities currently being played.
func FindMostPopularActivities(platform string) (*skillserver.EchoResponse, error) {
	response := skillserver.NewEchoResponse()

	translatedPlatform := platformNameToMapKey[platform]

	activityResp, err := client.GetPlayerActivity()
	if err != nil || activityResp == nil {
		response.OutputSpeech("Sorry Guardian, there was a problem contacting Charlemagne, " +
			"please try again later.")
		return response, nil
	}

	var activity map[string]*ActivitySummary
	if translatedPlatform == "" {
		activity = activityResp.ActivityByMode
	} else {
		activity = activityResp.ActivityByPlatform[translatedPlatform].ActivityByMode
	}
	fmt.Println(activity)
	ordered := sortPlayerActivityModes(activity)

	response.OutputSpeech("Guardian, according to Charlemagne, the top three activities " +
		"being played right now are: ")
	for _, activity := range ordered[:3] {
		response.OutputSpeech(modeTypeToName[activity.Mode])
	}
	response.OutputSpeech(". Go get that loot!")

	return response, nil
}

type activitySort []*ActivitySummary

func (activities activitySort) Len() int { return len(activities) }
func (activities activitySort) Swap(i, j int) {
	activities[i], activities[j] = activities[j], activities[i]
}
func (activities activitySort) Less(i, j int) bool {
	return activities[i].PercentagePlayed < activities[j].PercentagePlayed
}

func sortPlayerActivityModes(activity map[string]*ActivitySummary) []*ActivitySummary {

	result := make([]*ActivitySummary, 0, len(activity))

	for _, summary := range activity {
		result = append(result, summary)
	}

	sort.Sort(activitySort(result))
	return result
}
