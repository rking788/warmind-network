package bungie

import (
	"time"
)

// CharacterActivities represents the current and available activities for a particular
// character.
type CharacterActivities struct {
	DateActivityStarted         time.Time `json:"dateActivityStarted"`
	CurrentActivityHash         uint      `json:"currentActivityHash"`
	CurrentPlaylistActivityHash uint      `json:"currentPlaylistActivityHash"`
	CurrentActivityModeHash     uint      `json:"currentActivityModeHash"`
	CurrentActivityModeHashes   []uint    `json:"currentActivityModeHashes"`
	LastCompletedStoreHash      int       `json:"lastCompletedStoryHash"`

	// AvailableActivities []*AvailableActvities `json:"availableActivities"`
	// NOTE: This is likely a duplicate of the data in currentActivityModeHashes, use that instead
	// CurrentActivityModeTypes `json:"currentActivityModeTypes"`
	*Character
}

// ActivityStartSort is used to sort a slice of CharacterActivities by their activity start date.
// The list will be sorted from most recent to least recent start date.
type ActivityStartSort []*CharacterActivities

func (activities ActivityStartSort) Len() int { return len(activities) }
func (activities ActivityStartSort) Swap(i, j int) {
	activities[i], activities[j] = activities[j], activities[i]
}
func (activities ActivityStartSort) Less(i, j int) bool {
	return activities[i].DateActivityStarted.Before(activities[j].DateActivityStarted)
}
