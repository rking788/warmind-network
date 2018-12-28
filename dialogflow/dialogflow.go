package dialogflow

import (
	"strings"

	raven "github.com/getsentry/raven-go"
	"github.com/kpango/glg"
	"github.com/rking788/go-alexa/skillserver"
	"github.com/rking788/warmind-network/bungie"
	"github.com/rking788/warmind-network/db"
	df2 "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

type DialogFlowResponse struct {
	Payload *GooglePayload `json:"payload"`
}

type GooglePayload struct {
	Google *AssistantResponse `json:"google"`
}

type AssistantResponse struct {
	ExpectUserResponse bool `json:"expectUserResponse"`
	//Final              *FinalResponse `json:"finalResponse"`
	Rich *RichResponse `json:"richResponse"`
}

type AssistantResponseItem struct {
	Simple *SimpleResponse `json:"simpleResponse"`
}

type SimpleResponse struct {
	TextToSpeech string `json:"textToSpeech"`
	SSML         string `json:"ssml"`
	DisplayText  string `json:"displayText"`
}

type FinalResponse struct {
	Rich *RichResponse `json:"richResponse"`
}

type RichResponse struct {
	Items []*AssistantResponseItem `json:"items"`
}

func newGoogleDialogflowResponse() *DialogFlowResponse {
	response := &DialogFlowResponse{
		Payload: &GooglePayload{
			Google: &AssistantResponse{
				ExpectUserResponse: false,
				Rich:               &RichResponse{},
			},
		},
	}

	response.Payload.Google.Rich.Items = make([]*AssistantResponseItem, 0, 3)

	return response
}

func (r *DialogFlowResponse) setExpectUserResponse(expect bool) {
	r.Payload.Google.ExpectUserResponse = expect
}

func (r *DialogFlowResponse) setGoogleTextToSpeech(text string) {

	if len(r.Payload.Google.Rich.Items) == 0 {
		r.Payload.Google.Rich.Items = append(r.Payload.Google.Rich.Items, &AssistantResponseItem{Simple: &SimpleResponse{}})
	}

	r.Payload.Google.Rich.Items[0].Simple.TextToSpeech = text
}

func (r *DialogFlowResponse) setGoogleSSML(text string) {

	if len(r.Payload.Google.Rich.Items) == 0 {
		r.Payload.Google.Rich.Items = append(r.Payload.Google.Rich.Items, &AssistantResponseItem{Simple: &SimpleResponse{}})
	}

	r.Payload.Google.Rich.Items[0].Simple.SSML = text
}

func (r *DialogFlowResponse) setGoogleDisplayText(text string) {

	if len(r.Payload.Google.Rich.Items) == 0 {
		r.Payload.Google.Rich.Items = append(r.Payload.Google.Rich.Items, &AssistantResponseItem{Simple: &SimpleResponse{}})
	}

	r.Payload.Google.Rich.Items[0].Simple.DisplayText = text
}

func accessTokenFromRequest(r *df2.WebhookRequest) string {
	return r.GetOriginalDetectIntentRequest().GetPayload().GetFields()["user"].GetStructValue().GetFields()["accessToken"].GetStringValue()
}

// CountItem calls the Bungie API to see count the number of Items on all characters and
// in the vault.
func CountItem(r *df2.WebhookRequest) *DialogFlowResponse {

	response := newGoogleDialogflowResponse()
	accessToken := accessTokenFromRequest(r)

	// TODO: Get the item name slot value from the google request
	//item, _ := echoRequest.GetSlotValue("Item")
	item := "legendary shards"
	lowerItem := strings.ToLower(item)
	speech, err := bungie.CountItem(lowerItem, accessToken)
	if err != nil || speech == "" {
		raven.CaptureError(err, nil)
		glg.Errorf("Error counting the number of items: %s", err.Error())
		response.setGoogleTextToSpeech("Sorry Guardian, an error occurred counting that item.")
		return response
	}

	response.setGoogleTextToSpeech(speech)

	return response
}

// MaxPower will equip the loadout on the current character that provides the maximum
// amount of power.
func MaxPower(r *df2.WebhookRequest) *DialogFlowResponse {

	accessToken := accessTokenFromRequest(r)
	speech, err := bungie.EquipMaxLightGear(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping max power: %s", err.Error())
		speech = "Sorry Guardian, an error occurred equipping your max power gear."
	}

	response := newGoogleDialogflowResponse()
	response.setGoogleTextToSpeech(speech)

	return response
}

// GetCurrentRank will summarize the player's current ranking in various activities. The activity
// will be determined by a slot in the echo request. This could be either crucible for their
// glory and valor rankings or gambit for their infamy.
func GetCurrentRank(r *df2.WebhookRequest) *DialogFlowResponse {

	accessToken := accessTokenFromRequest(r)
	// TODO: Get slot value for progression name
	progression := "valor"
	//progression, _ := request.GetSlotValue("Progression")

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

	response := newGoogleDialogflowResponse()
	if speech == "" || err != nil {
		raven.CaptureError(err, nil)
		outputStr := "Sorry Guardian, there was an error getting your current " +
			progression + " ranking, please try again later."
		response.setGoogleTextToSpeech(outputStr)
		return response
	}

	response.setGoogleTextToSpeech(speech)
	return response
}

// EquipNamedLoadout will take the name of a loadout and try to retrieve it from the database
// and equip it on the user's currently active character.
func EquipNamedLoadout(r *df2.WebhookRequest) *DialogFlowResponse {

	response := newGoogleDialogflowResponse()
	accessToken := accessTokenFromRequest(r)

	// TODO: Get loadout name from slot values
	//loadoutName, _ := request.GetSlotValue("Name")
	loadoutName := "crucible"
	if loadoutName == "" {
		response.setGoogleTextToSpeech("Sorry Guardian, you must specify a name for the loadout being equipped.")
		return response
	}

	speech, err := bungie.EquipNamedLoadout(accessToken, loadoutName)

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error occurred equipping loadout: %s", err.Error())
		response.setGoogleTextToSpeech("Sorry Guardian, an error occurred equipping your loadout.")
		return response
	}

	response.setGoogleTextToSpeech(speech)
	return response
}

// ListLoadouts provides a speech response back to the user that lists the names
// of the currently saved loadouts that are found in the DB.
func ListLoadouts(r *df2.WebhookRequest) *DialogFlowResponse {

	response := newGoogleDialogflowResponse()
	accessToken := accessTokenFromRequest(r)
	speech, err := bungie.GetLoadoutNames(accessToken)
	if err != nil {
		raven.CaptureError(err, nil)
		response.setGoogleTextToSpeech("Sorry Guardian, there was an error getting your loadout names, please try again later")
		return response
	}

	response.setGoogleTextToSpeech(speech)
	return response
}

// DestinyJoke will return the desired text for a random joke from the database.
func DestinyJoke(r *df2.WebhookRequest) *DialogFlowResponse {

	response := newGoogleDialogflowResponse()
	setup, punchline, err := db.GetRandomJoke()
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error loading joke from DB: %s", err.Error())
		response.setGoogleTextToSpeech("Sorry Guardian, I was unable to load a joke right now.")
		return response
	}

	builder := skillserver.NewSSMLTextBuilder()
	builder.AppendPlainSpeech(setup).
		AppendBreak("2s", "medium", "").
		AppendPlainSpeech(punchline)

	// TODO: Is the syntax for the Google and Alexa SSML the same?
	response.setGoogleSSML(builder.Build())

	return response
}
