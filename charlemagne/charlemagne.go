package charlemagne

import (
	"errors"
	"time"

	"github.com/kpango/glg"
)

func InitEnv() {
	cachedActivityByPlatform = make(map[string][]*ActivitySummary)
	cachedMetaResponses = make(map[string]map[string]*MetaResponse)
	for _, platform := range []string{"", "1", "2", "4"} {
		cachedMetaResponses[platform] = make(map[string]*MetaResponse)
	}
	startCachePopulation()
}

func startCachePopulation() {

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			activityResponse, err := requestCurrentPlayerActivity()
			if err == nil {
				cachedActivityByMode = sortPlayerActivityModes(activityResponse.ActivityByMode)
				for platform, activity := range activityResponse.ActivityByPlatform {
					cachedActivityByPlatform[platform] = sortPlayerActivityModes(activity.ActivityByMode)
				}
			} else {
				glg.Warnf("Failed to cache player activity from Charlemagne")
			}

			// All platforms, xbox, psn, pc
			for _, platform := range []string{"", "1", "2", "4"} {
				for _, mode := range uniqueModeTypes() {
					cachedMetaResponses[platform][mode], err = requestCurrentMetaForMode(mode, platform)
					if err == nil {
                        // Success
					} else {
						glg.Warnf("Failed to request cached meta for mode=%s and platform=%s", mode, platform)
					}
				}
			}

			<-ticker.C
		}
	}()
}

func requestCurrentPlayerActivity() (*ActivityResponse, error) {
	activityResp, err := client.GetPlayerActivity()
	if err != nil || activityResp == nil {
		return nil, errors.New("Failed to load player activity from Charlemagne")
	}

	return activityResp, nil
}

func requestCurrentMetaForMode(mode, platform string) (*MetaResponse, error) {
	meta, err := client.GetCurrentMeta("", []string{mode}, platform)
	if err != nil {
		return nil, err
	}

	return meta, nil
}
