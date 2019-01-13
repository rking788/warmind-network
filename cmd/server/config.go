package main

import (
	"encoding/json"
	"os"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
)

// EnvConfig specifies all of the configuration that needs to be setup on different hosts or
// for different environments. This includes things like log leve, SSL config, Redis,
// and the Database which stores the Destiny manifest.
type EnvConfig struct {
	Environment              string `json:"environment"`
	RedisURL                 string `json:"redis_url"`
	BungieAPIKey             string `json:"bungie_api_key"`
	DatabaseURL              string `json:"database_url"`
	AlexaAppID               string `json:"alexa_app_id"`
	WarmindNetworkAlexaAppID string `json:"warmind_network_alexa_app_id"`
	WarmindBungieAPIKey      string `json:"warmind_bungie_api_key"`
	LogLevel                 string `json:"log_level"`
	LogFilePath              string `json:"log_file_path"`
	SSLCertPath              string `json:"ssl_cert_path"`
	SSLKeyPath               string `json:"ssl_key_path"`
	Port                     string `json:"port"`
	SentryDSN                string `json:"sentry_dsn"`
}

// NewEnvConfig will create a default instance of the EnvConfig struct
func NewEnvConfig() *EnvConfig {
	// Default to values from the environment or nothing, this is mainly for the Heroku deployments
	config := &EnvConfig{
		Environment:              "staging",
		RedisURL:                 os.Getenv("REDIS_URL"),
		BungieAPIKey:             os.Getenv("BUNGIE_API_KEY"),
		WarmindBungieAPIKey:      os.Getenv("WARMIND_BUNGIE_API_KEY"),
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		AlexaAppID:               os.Getenv("ALEXA_APP_ID"),
		WarmindNetworkAlexaAppID: os.Getenv("WARMIND_NETWORK_APP_ID"),
		LogLevel:                 os.Getenv("GUARDIAN_HELPER_LOG_LEVEL"),
		Port:                     os.Getenv("PORT"),
		SentryDSN:                os.Getenv("SENTRY_DSN"),
	}

	return config
}

func loadConfig(path *string) (config *EnvConfig) {
	config = NewEnvConfig()
	if *path == "" {
		return
	}

	in, err := os.Open(*path)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error trying to open the specified config file: %s", err.Error())
		return
	}

	err = json.NewDecoder(in).Decode(&config)
	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Error deserializing config JSON: %s", err.Error())
		return
	}

	return
}
