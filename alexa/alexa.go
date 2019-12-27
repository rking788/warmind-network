package alexa

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/getsentry/raven-go"
	"github.com/kpango/glg"
	"github.com/mikeflynn/go-alexa/skillserver"
	"github.com/mikeflynn/go-alexa/skillserver/dialog"
	"github.com/rking788/warmind-network/bungie"
	"github.com/rking788/warmind-network/charlemagne"
	"github.com/rking788/warmind-network/db"
	"github.com/rking788/warmind-network/trials"
)

type DialogType string

const (
	DialogDelegate      DialogType = "Dialog.Delegate"
	DialogElicitSlot    DialogType = "Dialog.ElicitSlot"
	DialogConfirmSlot   DialogType = "Dialog.ConfirmSlot"
	DialogConfirmIntent DialogType = "Dialog.ConfirmIntent"
)

type DialogState string

const (
	DialogStarted    DialogState = "STARTED"
	DialogInProgress DialogState = "IN_PROGRESS"
	DialogCompleted  DialogState = "COMPLETED"
)

type ConfirmationStatus string

const (
	ConfirmationNone      ConfirmationStatus = "NONE"
	ConfirmationConfirmed ConfirmationStatus = "CONFIRMED"
	ConfirmationDenied    ConfirmationStatus = "DENIED"
)

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

var redisConnPool *redis.Pool

// InitEnv provides a package level initialization point for any work that is environment specific
func InitEnv(redisURL string) {
	redisConnPool = newRedisPool(redisURL)
}

// Redis related functions

func newRedisPool(addr string) *redis.Pool {
	// 25 is the maximum number of active connections for the Heroku Redis free tier
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   25,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.DialURL(addr) },
	}
}

// GetSession will attempt to read a session from the cache, if an existing one is not found, an empty session
// will be created with the specified sessionID.
func GetSession(sessionID string) (session *Session) {
	session = &Session{ID: sessionID}

	conn := redisConnPool.Get()
	defer conn.Close()

	key := fmt.Sprintf("sessions:%s", sessionID)
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// NOTE: This is a normal situation, if the session is not stored in the cache, it will hit this condition.
		return
	}

	err = json.Unmarshal([]byte(reply), session)

	return
}

// SaveSession will persist the given session to the cache. This will allow support for long running
// Alexa sessions that continually prompt the user for more information.
func SaveSession(session *Session) {

	conn := redisConnPool.Get()
	defer conn.Close()

	sessionBytes, err := json.Marshal(session)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Couldn't marshal session to string: %s", err.Error())
		return
	}

	key := fmt.Sprintf("sessions:%s", session.ID)
	_, err = conn.Do("SET", key, string(sessionBytes))
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to set session: %s", err.Error())
	}
}

// ClearSession will remove the specified session from the local cache, this will be done
// when the user completes a full request session.
func ClearSession(sessionID string) {

	conn := redisConnPool.Get()
	defer conn.Close()

	key := fmt.Sprintf("sessions:%s", sessionID)
	_, err := conn.Do("DEL", key)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to delete the session from the Redis cache: %s", err.Error())
	}
}

// Handler is the type of function that should be used to respond to a specific intent.
type Handler func(*skillserver.EchoRequest) *skillserver.EchoResponse

// AuthWrapper is a handler function wrapper that will fail the chain of handlers if an access token was not provided
// as part of the Alexa request
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
func CountItem(echoRequest *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := echoRequest.Session.User.AccessToken
	item, _ := echoRequest.GetSlotValue("Item")
	lowerItem := strings.ToLower(item)
	speech, err := bungie.CountItem(lowerItem, accessToken)
	if err != nil || speech == "" {
		raven.CaptureError(err, nil)
		glg.Errorf("Error counting the number of items: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred counting that item.")
		return response
	}

	response.OutputSpeech(speech)

	return response
}

// TransferItem will attempt to transfer either a specific quantity or all of a
// specific item to a specified character. The item name and destination are the
// required fields. The quantity and source are optional.
func TransferItem(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	countStr, _ := request.GetSlotValue("Count")
	count := -1
	if countStr != "" {
		tempCount, ok := strconv.Atoi(countStr)
		if ok != nil {
			response = skillserver.NewEchoResponse()
			response.OutputSpeech("Sorry Guardian, I didn't understand the number you asked to be transferred. Do not specify a quantity if you want all to be transferred.")
			return response
		}

		if tempCount <= 0 {
			output := fmt.Sprintf("Sorry Guardian, you need to specify a positive, non-zero number to be transferred, not %d", tempCount)
			response.OutputSpeech(output)
			return response
		}

		count = tempCount
	}

	item, _ := request.GetSlotValue("Item")
	sourceClass, _ := request.GetSlotValue("Source")
	destinationClass, _ := request.GetSlotValue("Destination")
	if destinationClass == "" {
		response.OutputSpeech("Sorry Guardian, you must specify a destination for the items to be transferred.")
		return response
	}

	glg.Infof("Transferring %d of your %s from your %s to your %s", count, strings.ToLower(item), strings.ToLower(sourceClass), strings.ToLower(destinationClass))
	speech, err := bungie.TransferItem(strings.ToLower(item), accessToken, strings.ToLower(sourceClass), strings.ToLower(destinationClass), count)
	if err != nil {
		raven.CaptureError(err, nil)
		response.OutputSpeech("Sorry Guardian, an error occurred trying to transfer that item.")
		return response
	}

	response.OutputSpeech(speech)
	return response
}

// MaxPower will equip the loadout on the current character that provides the maximum
// amount of power.
func MaxPower(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	speech, err := bungie.EquipMaxLightGear(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping max power: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred equipping your max power gear.")
	}

	response.OutputSpeech(speech)
	return response
}

// RandomGear will equip random gear for each of the current characters gear slots (excluding
// things like ghost and class items). It is possible to randomize only weapons or weapons
// and armor.
func RandomGear(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	speech, err := bungie.RandomizeLoadout(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping random gear: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred equipping random gear.")
	}

	response.OutputSpeech(speech)
	return response
}

// UnloadEngrams will take all engrams on all of the current user's characters and
// transfer them all to the vault to allow the player to continue farming.
func UnloadEngrams(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	speech, err := bungie.UnloadEngrams(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred unloading engrams: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred moving your engrams.")
		return response
	}

	response.OutputSpeech(speech)
	return response
}

// DestinyJoke will return the desired text for a random joke from the database.
func DestinyJoke(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	setup, punchline, err := db.GetRandomJoke()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error loading joke from DB: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, I was unable to load a joke right now.")
		return response
	}

	builder := skillserver.NewSSMLTextBuilder()
	builder.AppendPlainSpeech(setup).
		AppendBreak("medium", "2s").
		AppendPlainSpeech(punchline)
	response.OutputSpeechSSML(builder.Build())

	return response
}

// CreateLoadout will determine the user's current character and create a new loadout based on
// their currently equipped items. This loadout is then serialized and persisted to a
// persistent store.
func CreateLoadout(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()

	glg.Debugf("Found dialog state = %s", request.Request.DialogState)
	if DialogState(request.Request.DialogState) != DialogCompleted {
		// The user still needs to provide a name for the new loadout to be created
		response = response.RespondToIntent(dialog.Delegate, nil, nil)
		response.EndSession(false)
		return response
	}

	glg.Debugf("Found intent confirmation status = %s", request.Request.Intent.ConfirmationStatus)
	intentConfirmation := request.Request.Intent.ConfirmationStatus
	if ConfirmationStatus(intentConfirmation) == ConfirmationDenied {
		// The user does NOT want to overwrite the existing loadout with the same name
		return response
	}

	accessToken := request.Session.User.AccessToken
	loadoutName, _ := request.GetSlotValue("Name")
	if loadoutName == "" {
		response.OutputSpeech("Sorry Guardian, you must specify a name for the loadout being saved.")
		return response
	}

	exists, _ := bungie.DoesLoadoutExist(accessToken, loadoutName)
	if exists && ConfirmationStatus(intentConfirmation) == ConfirmationNone {
		// Ask the user if they wish to overwrite this loadout
		response = response.RespondToIntent(dialog.ConfirmIntent, &request.Request.Intent, nil).
			OutputSpeech(fmt.Sprintf("You already have a loadout named %s, would you like to overwrite it?", loadoutName))
		response.EndSession(false)
		return response
	}

	var err error
	shouldOverwrite := ConfirmationStatus(intentConfirmation) == ConfirmationConfirmed
	speech, err := bungie.CreateLoadoutForCurrentCharacter(accessToken, loadoutName, shouldOverwrite)

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred creating loadout: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred saving your loadout.")
	}

	response.OutputSpeech(speech)

	return response
}

// EquipNamedLoadout will take the name of a loadout and try to retrieve it from the database
// and equip it on the user's currently active character.
func EquipNamedLoadout(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	loadoutName, _ := request.GetSlotValue("Name")
	if loadoutName == "" {
		response.OutputSpeech("Sorry Guardian, you must specify a name for the loadout being equipped.")
		return response
	}

	speech, err := bungie.EquipNamedLoadout(accessToken, loadoutName)

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping loadout: %s", err.Error())
		response.OutputSpeech("Sorry Guardian, an error occurred equipping your loadout.")
		return response
	}

	response.OutputSpeech(speech)
	return response
}

// ListLoadouts provides a speech response back to the user that lists the names
// of the currently saved loadouts that are found in the DB.
func ListLoadouts(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	response := skillserver.NewEchoResponse()
	accessToken := request.Session.User.AccessToken
	speech, err := bungie.GetLoadoutNames(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, there was an error getting your loadout names, please try again later")
		return response
	}

	response.OutputSpeech(speech)
	return response
}

// GetCurrentRank will summarize the player's current ranking in various activities. The activity
// will be determined by a slot in the echo request. This could be either crucible for their
// glory and valor rankings or gambit for their infamy.
func GetCurrentRank(request *skillserver.EchoRequest) *skillserver.EchoResponse {

	accessToken := request.Session.User.AccessToken
	progression, _ := request.GetSlotValue("Progression")

	var speech string
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
		speech, err = bungie.GetCurrentCrucibleRanking(accessToken)
	case "infamy":
		fallthrough
	case "gambit":
		speech, err = bungie.GetCurrentGambitRanking(accessToken)
	default:
		speech = "Sorry Guardian, I'm not sure how to get information about a " + progression + " rank"
	}

	response := skillserver.NewEchoResponse()
	if speech == "" || err != nil {
		raven.CaptureError(err, nil)
		outputStr := "Sorry Guardian, there was an error getting your current " +
			progression + " ranking, please try again later."
		response.OutputSpeech(outputStr)
		return response
	}

	response.OutputSpeech(speech)
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
