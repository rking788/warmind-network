package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/rking788/warmind-network/charlemagne"
	"github.com/rking788/warmind-network/db"
	"github.com/rking788/warmind-network/trials"

	"github.com/kpango/glg"
	"github.com/rking788/warmind-network/alexa"
	"github.com/rking788/warmind-network/bungie"

	"github.com/rking788/go-alexa/skillserver"
)

// AlexaHandlers are the handler functions mapped by the intent name that they should handle.
var (
	AlexaHandlers = map[string]alexa.Handler{
		"CountItem":                alexa.AuthWrapper(alexa.CountItem),
		"TransferItem":             alexa.AuthWrapper(alexa.TransferItem),
		"TrialsCurrentMap":         alexa.CurrentTrialsMap,
		"TrialsCurrentWeek":        alexa.AuthWrapper(alexa.CurrentTrialsWeek),
		"TrialsPersonalTopWeapons": alexa.AuthWrapper(alexa.PersonalTopWeapons),
		"UnloadEngrams":            alexa.AuthWrapper(alexa.UnloadEngrams),
		"EquipMaxLight":            alexa.AuthWrapper(alexa.MaxPower),
		"RandomizeGear":            alexa.AuthWrapper(alexa.RandomGear),
		"DestinyJoke":              alexa.DestinyJoke,
		"CreateLoadout":            alexa.AuthWrapper(alexa.CreateLoadout),
		"EquipNamedLoadout":        alexa.AuthWrapper(alexa.EquipNamedLoadout),
		"ListLoadouts":             alexa.AuthWrapper(alexa.ListLoadouts),
		"TopActivities":            alexa.TopActivities,
		"CurrentMeta":              alexa.CurrentMeta,
		"CrucibleRank":             alexa.AuthWrapper(alexa.GetCurrentRank),
		"CurrentRank":              alexa.AuthWrapper(alexa.GetCurrentRank),
		"AMAZON.HelpIntent":        alexa.HelpPrompt,
	}
)

var configPath = flag.String("config", "", "path to the environment configuration file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

// Applications is a definition of the Alexa applications running on this server.
var applications map[string]interface{}

// config is the environment configuration for this specific deployment of the server
var config *EnvConfig

// InitEnv is responsible for initializing all components (including sub-packages) that
// depend on a specific deployment environment configuration.
func InitEnv(c *EnvConfig) {
	applications = map[string]interface{}{
		"/echo/guardian-helper": skillserver.EchoApplication{ // Route
			AppID:          c.AlexaAppID, // Echo App ID from Amazon Dashboard
			OnIntent:       guardianHelperIntentHandler,
			OnLaunch:       guardianHelperIntentHandler,
			OnSessionEnded: EchoSessionEndedHandler,
		},
		"/echo/warmind-network": skillserver.EchoApplication{ // Route
			AppID:          c.WarmindNetworkAlexaAppID, // Echo App ID from Amazon Dashboard
			OnIntent:       warmindIntentHandler,
			OnLaunch:       warmindIntentHandler,
			OnSessionEnded: EchoSessionEndedHandler,
		},
		"/health": skillserver.StdApplication{
			Methods: "GET",
			Handler: healthHandler,
		},
	}

	ConfigureLogging(c.LogLevel, c.LogFilePath)
	raven.SetDSN(c.SentryDSN)

	// This provides and explicit configuration point as opposed to the package level
	// init functions, as well as making it easier to write unit tests.
	// It also makes it easier to guarantee ordering if that is necessary.
	trials.InitEnv(c.BungieAPIKey, c.WarmindBungieAPIKey, "")
	db.InitEnv(c.DatabaseURL)
	alexa.InitEnv(c.RedisURL)
	bungie.InitEnv(c.BungieAPIKey, c.WarmindBungieAPIKey)
	charlemagne.InitEnv()
}

func main() {

	flag.Parse()

	config = loadConfig(configPath)

	glg.Infof("Loaded config : %+v\n", config)
	InitEnv(config)

	defer CloseLogger()

	glg.Printf("Version=%s, BuildDate=%v", Version, BuildDate)

	// writeHeapProfile()

	if config.Environment == "production" {
		port := ":" + config.Port
		err := skillserver.RunSSL(applications, port, config.SSLCertPath, config.SSLKeyPath)
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Error starting the application! : %s", err.Error())
		}
	} else {
		// Heroku makes us read a random port from the environment and our app is a
		// subdomain of theirs so we get SSL for free
		skillserver.Run(applications, config.Port)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Up"))
}

func writeHeapProfile() {
	//bungie.EquipMaxLightGear("access-token")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if *memprofile != "" {
				f, err := os.Create(*memprofile)
				if err != nil {
					glg.Fatal(err)
				}
				pprof.WriteHeapProfile(f)
				f.Close()
				os.Exit(1)
				return
			}
		}
	}()
}

// Alexa skill related functions

// EchoSessionEndedHandler is responsible for cleaning up an open session since the user
// has quit the session.
func EchoSessionEndedHandler(echoRequest *skillserver.EchoRequest, echoResponse *skillserver.EchoResponse) {
	*echoResponse = *skillserver.NewEchoResponse()

	alexa.ClearSession(echoRequest.GetSessionID())
}

func guardianHelperIntentHandler(echoRequest *skillserver.EchoRequest, echoResponse *skillserver.EchoResponse) {
	glg.Info("Breaking the bad news to guardian-helper user")
	response := skillserver.NewEchoResponse()
	response.OutputSpeech("Sorry Guardian, the Guardian Helper skill is now shutdown, " +
		"please enable the Warmind Network skill for new features and " +
		" performance improvements.")
	*echoResponse = *response
	return
}

func warmindIntentHandler(echoRequest *skillserver.EchoRequest, echoResponse *skillserver.EchoResponse) {
	EchoIntentHandler(echoRequest, echoResponse)
}

// EchoIntentHandler is a handler method that is responsible for receiving the
// call from a Alexa command and returning the correct speech or cards.
func EchoIntentHandler(echoRequest *skillserver.EchoRequest, echoResponse *skillserver.EchoResponse) {

	// Time the intent handler to determine if it is taking longer than normal
	startTime := time.Now()
	defer func(start time.Time) {
		glg.Successf("IntentHandler execution time: %v", time.Since(start))
	}(startTime)

	var response *skillserver.EchoResponse

	// See if there is an existing session, or create a new one.
	session := alexa.GetSession(echoRequest.GetSessionID())
	alexa.SaveSession(session)

	intentName := echoRequest.GetIntentName()

	glg.Infof("RequestType: %s, IntentName: %s", echoRequest.GetRequestType(), intentName)

	handler, ok := AlexaHandlers[intentName]
	if echoRequest.GetRequestType() == "LaunchRequest" {
		response = alexa.WelcomePrompt(echoRequest)
	} else if intentName == "AMAZON.StopIntent" {
		response = skillserver.NewEchoResponse()
	} else if intentName == "AMAZON.CancelIntent" {
		response = skillserver.NewEchoResponse()
	} else if ok {
		response = handler(echoRequest)
	} else {
		response = skillserver.NewEchoResponse()
		response.OutputSpeech("Sorry Guardian, I did not understand your request.")
	}

	if response.Response.ShouldEndSession {
		alexa.ClearSession(session.ID)
	}

	*echoResponse = *response
}

// func dumpRequest(ctx *gin.Context) {

// 	data, err := httputil.DumpRequest(ctx.Request, true)
// 	if err != nil {
// 		glg.Errorf("Failed to dump the request: %s", err.Error())
// 		return
// 	}

// 	glg.Debug(string(data))
// }
