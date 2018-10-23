package models

import "time"

// CharacterActivities represents the current and available activities for a particular
// character. This is used to represent the raw JSON representation returned by the Bungie API
// the activity/playlist/mode hashes can be used to retrieve more details from the manifest out of the database.
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

type activityBase struct {
	Hash        uint
	Name        string
	Description string
}

// ActivityModifier represents a single modifier attached to an activity.
// (Example: Solar singe, Blackout, Glass, etc.)
type ActivityModifier struct {
	*activityBase
}

// ActivityType represents the type of a particular activity mostly defined in the Bungie manifest.
// This could be things like supremacy, mayhem, iron banner, story.
type ActivityType struct {
	*activityBase
}

// ActivityMode is the mode of the activity it is attached to. Similar to the activity type but with more
// data.
type ActivityMode struct {
	*activityBase
	ModeType    int
	Category    int
	Tier        int
	IsAggregate bool
	IsTeamBased bool
}

// Place is a representation of a specific location for an activity.
// Example: The Tangled Shore, Earth, The Crucible, "Titan, Moon of Saturn", etc.
type Place struct {
	*activityBase
}

// Destination represents a more detailed version of a Place. Destinations also contain
// information about bubbles.
// Example: The Farm, Arcadian Valley, Hellas Basin, The Tangled Shore
type Destination struct {
	*activityBase
	PlaceHash uint
}

// ActivityReward describes a possible item reward from a given activity as well as the expected quantity.
type ActivityReward struct {
	ItemHash uint
	Quantity uint
}

// ActivityChallenge represents a challenge that could be active for a particular activity.
// ObjectiveHash should be used to determine the goal of the challenge.
type ActivityChallenge struct {
	RewardSiteHash           uint
	InhibitRewardsUnlockHash uint
	ObjectiveHash            uint
}

// ActivityPlaylistItem describes an item in the playlist represented by the current activity.
type ActivityPlaylistItem struct {
	ActivityHash     uint `json:"activityHash"`
	ActivityModeHash uint `json:"activityModeHash"`
	Weight           uint `json:"weight"`
}

// ActivityMatchmaking describes the different matchmaking parameters for a particular activity.
type ActivityMatchmaking struct {
	IsMatchmade bool `json:"isMatchmade"`
	MinParty    int  `json:"minParty"`
	MaxParty    int  `json:"maxParty"`
	MaxPlayers  int  `json:"maxPlayers"`
	// TODO: Not sure wth this is
	RequiresGuardianOath bool `json:"requiresGuardianOath"`
}

// Activity is the top level item that can represent a current or completed activity for a player.
// This type will contain information about modifiers, challenges, the mode or type of the activity,
// places, and destinations.
type Activity struct {
	*activityBase
	LightLevel             uint
	DestinationHash        *Destination
	PlaceHash              *Place
	ActivityTypeHash       *ActivityType
	IsPlaylist             bool
	IsPVP                  bool
	DirectActivityModeHash uint
	DirectActivityModeType int
	ActivityModeHashes     []uint // I think the hashes here work well, ActivityMode is probably unnecessary
	ActivityModeTypes      []int

	Rewards       []*ActivityReward
	Modifiers     []uint // I think the hashes are good for this, ActivityModifier is probably unnecessary
	Challenges    []*ActivityChallenge
	PlaylistItems []*ActivityPlaylistItem

	Matchmaking *ActivityMatchmaking
}
