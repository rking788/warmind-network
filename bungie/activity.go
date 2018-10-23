package bungie

import "github.com/rking788/warmind-network/models"

// ActivityStartSort is used to sort a slice of CharacterActivities by their activity start date.
// The list will be sorted from most recent to least recent start date.
type ActivityStartSort []*models.CharacterActivities

func (activities ActivityStartSort) Len() int { return len(activities) }
func (activities ActivityStartSort) Swap(i, j int) {
	activities[i], activities[j] = activities[j], activities[i]
}
func (activities ActivityStartSort) Less(i, j int) bool {
	return activities[i].DateActivityStarted.Before(activities[j].DateActivityStarted)
}
