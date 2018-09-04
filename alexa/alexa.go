package alexa

import (
	"fmt"
	"strconv"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
	"github.com/rking788/warmind-network/bungie"
	"github.com/rking788/warmind-network/charlemagne"
	"github.com/rking788/warmind-network/storage"
	"github.com/rking788/warmind-network/trials"

	"strings"

	"github.com/rking788/go-alexa/skillserver"
)

type SessionCache interface {
	GetSession(id string, session interface{}) error
	SaveSession(id string, session interface{})
	ClearSession(id string)
}

// Session is responsible for storing information related to a specific skill invocation.
// A session will remain open if the LaunchRequest was received.
type Session struct {
	ID                   string
	Action               string
	ItemName             string
	DestinationClassHash int
	SourceClassHash      int
	Quantity             int
}

var (
	cache SessionCache
)

// InitEnv provides a package level initialization point for any work that is environment specific
func InitEnv(sessionCache SessionCache) {
	cache = sessionCache
}

// Handler is the type of function that should be used to respond to a specific intent.
type Handler func(*skillserver.EchoRequest) *skillserver.EchoResponse

func GetSession(id string) *Session {
	s := &Session{ID: id}

	// At this point most of the errors will be from missing sessions in the cache (new sessions)
	// so ignore the warning here and let Sentry report the connection pool exhausted errors
	_ = cache.GetSession(id, s)

	return s
}

func SaveSession(session *Session) {
	cache.SaveSession(session.ID, session)
}

func ClearSession(id string) {
	cache.ClearSession(id)
}

// AuthWrapper is a handler function wrapper that will fail the chain of handlers
// if an access token was not provided as part of the Alexa request
func AuthWrapper(handler Handler) Handler {

	return func(req *skillserver.EchoRequest) *skillserver.EchoResponse {
		accessToken := req.Session.User.AccessToken
		if accessToken == "" {
			response := skillserver.NewEchoResponse()
			response.
				OutputSpeech("Sorry Guardian, it looks like your Bungie.net account needs to be linked in the Alexa app.").
				LinkAccountCard()
			return response
		}

		return handler(req)
	}
}

// WelcomePrompt is responsible for prompting the user with information about what they can ask
// the skill to do.
func WelcomePrompt(echoRequest *skillserver.EchoRequest) (response *skillserver.EchoResponse) {
	response = skillserver.NewEchoResponse()

	response.OutputSpeech("Welcome Guardian, would you like to equip max power, or transfer an item to a specific character, " +
		"find out how many of an item you have, or ask about Trials of the Nine?").
		Reprompt("Do you want to equip max power, transfer an item, find out how much of an item you have, or ask about Trials of the Nine?").
		EndSession(false)

	return
}

// HelpPrompt provides the required information to satisfy the HelpIntent built-in Alexa intent.
// This should provider information to the user to let them know what the skill can do without
// providing exact commands.
func HelpPrompt(echoRequest *skillserver.EchoRequest) (response *skillserver.EchoResponse) {
	response = skillserver.NewEchoResponse()

	response.OutputSpeech("Welcome Guardian, I am here to help manage your Destiny in-game inventory. You can ask " +
		"me to equip your max power loadout, or transfer items between your available " +
		"characters including the vault. You can also ask how many of an " +
		"item you have. Trials of the Nine statistics provided by Trials Report are available too.").
		EndSession(false)

	return
}

// CountItem calls the Bungie API to see count the number of Items on all characters and
// in the vault.
func CountItem(echoRequest *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := echoRequest.Session.User.AccessToken
	item, _ := echoRequest.GetSlotValue("Item")
	lowerItem := strings.ToLower(item)
	response, err := bungie.CountItem(lowerItem, accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error counting the number of items: %s", err.Error())
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, an error occurred counting that item.")
	}

	return
}

// TransferItem will attempt to transfer either a specific quantity or all of a
// specific item to a specified character. The item name and destination are the
// required fields. The quantity and source are optional.
func TransferItem(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	countStr, _ := request.GetSlotValue("Count")
	count := -1
	if countStr != "" {
		tempCount, ok := strconv.Atoi(countStr)
		if ok != nil {
			response = skillserver.NewEchoResponse()
			response.OutputSpeech("Sorry Guardian, I didn't understand the number you asked to be transferred. Do not specify a quantity if you want all to be transferred.")
			return
		}

		if tempCount <= 0 {
			output := fmt.Sprintf("Sorry Guardian, you need to specify a positive, non-zero number to be transferred, not %d", tempCount)
			response.OutputSpeech(output)
			return
		}

		count = tempCount
	}

	item, _ := request.GetSlotValue("Item")
	sourceClass, _ := request.GetSlotValue("Source")
	destinationClass, _ := request.GetSlotValue("Destination")
	if destinationClass == "" {
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, you must specify a destination for the items to be transferred.")
		return
	}

	glg.Infof("Transferring %d of your %s from your %s to your %s", count, strings.ToLower(item), strings.ToLower(sourceClass), strings.ToLower(destinationClass))
	response, err := bungie.TransferItem(strings.ToLower(item), accessToken, strings.ToLower(sourceClass), strings.ToLower(destinationClass), count)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, an error occurred trying to transfer that item.")
		return
	}

	return
}

// MaxPower will equip the loadout on the current character that provides the maximum
// amount of power.
func MaxPower(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := bungie.EquipMaxLightGear(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping max power: %s", err.Error())
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, an error occurred equipping your max power gear.")
	}

	return
}

// RandomGear will equip random gear for each of the current characters gear slots (excluding
// things like ghost and class items). It is possible to randomize only weapons or weapons
// and armor.
func RandomGear(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := bungie.RandomizeLoadout(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping random gear: %s", err.Error())
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, an error occurred equipping random gear.")
	}

	return
}

// UnloadEngrams will take all engrams on all of the current user's characters and
// transfer them all to the vault to allow the player to continue farming.
func UnloadEngrams(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := bungie.UnloadEngrams(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred unloading engrams: %s", err.Error())
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, an error occurred moving your engrams.")
	}

	return
}

// DestinyJoke will return the desired text for a random joke from the database.
func DestinyJoke(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	response = skillserver.NewEchoResponse()
	setup, punchline, err := storage.GetRandomJoke()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error loading joke from DB: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, I was unable to load a joke right now.")
		return
	}

	builder := skillserver.NewSSMLTextBuilder()
	builder.AppendPlainSpeech(setup).
		AppendBreak("2s", "medium", "").
		AppendPlainSpeech(punchline)
	response.OutputSpeechSSML(builder.Build())

	return
}

// CreateLoadout will determine the user's current character and create a new loadout based on
// their currently equipped items. This loadout is then serialized and persisted to a
// persistent store.
func CreateLoadout(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	response = skillserver.NewEchoResponse()

	glg.Debugf("Found dialog state = %s", request.GetDialogState())
	if request.GetDialogState() != skillserver.DialogCompleted {
		// The user still needs to provide a name for the new loadout to be created
		response.DialogDelegate(nil)
		return
	}

	glg.Debugf("Found intent confirmation status = %s", request.GetIntentConfirmationStatus())
	intentConfirmation := request.GetIntentConfirmationStatus()
	if intentConfirmation == skillserver.ConfirmationDenied {
		// The user does NOT want to overwrite the existing loadout with the same name
		return
	}

	accessToken := request.Session.User.AccessToken
	loadoutName, _ := request.GetSlotValue("Name")
	if loadoutName == "" {
		response.OutputSpeech("Sorry Guardian, you must specify a name for the loadout being saved.")
	}

	var err error
	response, err = bungie.CreateLoadoutForCurrentCharacter(accessToken, loadoutName,
		intentConfirmation == "CONFIRMED")

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred creating loadout: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred saving your loadout.")
	}

	return
}

// EquipNamedLoadout will take the name of a loadout and try to retrieve it from the database
// and equip it on the user's currently active character.
func EquipNamedLoadout(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	response = skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	loadoutName, _ := request.GetSlotValue("Name")
	if loadoutName == "" {
		response.OutputSpeech("Sorry Guardian, you must specify a name for the loadout being equipped.")
	}

	response, err := bungie.EquipNamedLoadout(accessToken, loadoutName)

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping loadout: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred equipping your loadout.")
	}

	return
}

// ListLoadouts provides a speech response back to the user that lists the names
// of the currently saved loadouts that are found in the DB.
func ListLoadouts(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := bungie.GetLoadoutNames(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, there was an error getting your loadout names, please try again later")
		return
	}

	return
}

// GetCurrentRank will summarize the player's current ranking in various activities. The activity
// will be determined by a slot in the echo request. This could be either crucible for their
// glory and valor rankings or gambit for their infamy.
func GetCurrentRank(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	accessToken := request.Session.User.AccessToken
	progression, _ := request.GetSlotValue("Progression")

	var response *skillserver.EchoResponse
	var err error
	switch strings.ToLower(progression) {
	case "":
		fallthrough
	case "valor":
		fallthrough
	case "glory":
		fallthrough
	case "quickplay":
		fallthrough
	case "competitive":
		fallthrough
	case "crucible":
		response, err = bungie.GetCurrentCrucibleRanking(accessToken)
	case "infamy":
		fallthrough
	case "gambit":
		response, err = bungie.GetCurrentGambitRanking(accessToken)
	default:
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, I'm not sure how to get information about a " + progression + " rank")
	}

	if response == nil || err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, there was an error getting your current " +
			progression + " ranking, please try again later.")
		return response
	}

	return response
}

/**
 * Trials of the Nine
 */

// CurrentTrialsMap will return a brief description of the current map in the active
// Trials of the Nine week.
func CurrentTrialsMap(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	response, err := trials.GetCurrentMap()
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, I cannot access this information right now, please try again later.")
		return
	}

	return
}

// CurrentTrialsWeek will return a brief description of the current map in the
// active Trials of the Nine week.
func CurrentTrialsWeek(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := trials.GetCurrentWeek(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, I cannot access this information right now, " +
			"please try again later.")
		return
	}

	return
}

// PersonalTopWeapons will check Trials Report for the most used weapons for the current user.
func PersonalTopWeapons(request *skillserver.EchoRequest) (response *skillserver.EchoResponse) {

	accessToken := request.Session.User.AccessToken
	response, err := trials.GetPersonalTopWeapons(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, I cannot access this information at " +
			"this time, please try again later")
		return
	}

	return
}

/**
 * Charlemagne
 */

// TopActivities is responsible for describing to the user, the most popular activities currently
// being played in game. This could help users decide what activities have the best rewards
// or can be part of the "grind" for better gear.
func TopActivities(echoRequest *skillserver.EchoRequest) *skillserver.EchoResponse {

	platform, _ := echoRequest.GetSlotValue("Platform")

	response, err := charlemagne.FindMostPopularActivities(platform)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, there was an issue contacting Charlemagne, " +
			"please try again later.")
	}

	return response
}

// CurrentMeta will present data to the user from Charlemagne about which weapons are currently
// being used the most in game. This could be on a particular platform or in a particular
// game mode (PvP, nightfall, raid, etc.)
func CurrentMeta(echoRequest *skillserver.EchoRequest) *skillserver.EchoResponse {

	platform, _ := echoRequest.GetSlotValue("Platform")
	activity, _ := echoRequest.GetSlotValue("GameMode")

	platform = strings.ToLower(platform)
	activity = strings.ToLower(activity)

	glg.Warnf("Requesting current meta with game mode name: %s", activity)

	// If game mode is Trials related, use Trials Report data.
	if strings.Contains(activity, "trials") {
		response, err := trials.GetWeaponUsagePercentages()
		if err != nil {
			raven.CaptureError(err, nil)
			response = skillserver.NewEchoResponse()
			response.OutputSpeech("Sorry Guardian, I cannot access this information at this " +
				" time, please try again later")
		}

		return response
	}

	response, err := charlemagne.FindCurrentMeta(platform, activity)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, there was a problem contacting Charlemagne, " +
			" please try again later.")
	}

	return response
}
