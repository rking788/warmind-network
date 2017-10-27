package trials

// Constant Trials Report API endpoints
const (
	NineBaseURL              = "https://api.trialsofthenine.com"
	NineCurrentWeekStatsPath = "/week/0/"
	// MembershipID
	NinePlayerSummaryFmt              = "/player/%s/"
	NinePlayerCurrentWeekStatsPathFmt = "/slack/trials/week/%s/0"
	NineRivalsPathFmt                 = "/player/%s/rivals/"
	NineFireteamFmt                   = "/player/%s/fireteam/"
	NineTeammatesFmt                  = "/player/%s/teammates/"

	// How many weapons to return in the Alexa response describing usage stats
	TopWeaponUsageLimit = 3
)
