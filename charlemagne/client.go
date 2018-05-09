package charlemagne

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const (
	metaResourcePath     = "/meta"
	activityResourcePath = "/playerActivity"
)

var client = newClient("")

// Client is a type that encapsulates functionality for making requests
// to the Charlemagne API.
type Client struct {
	*http.Client
	BaseURL string
}

func newClient(baseURL string) *Client {

	base := "https://api.warmind.io"
	if baseURL != "" {
		base = baseURL
	}

	return &Client{
		Client:  http.DefaultClient,
		BaseURL: base,
	}
}

// GetPlayerActivity will send a request to the Charlemagne API to get the current activity
// for all players in the game.
func (c *Client) GetPlayerActivity() (*ActivityResponse, error) {

	reqURL := c.BaseURL + activityResourcePath
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	result := &ActivityResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(result)
	if err != nil {
		return nil, err
	}

	// The activity summary objects contain game modes as ints. Instead of having to do the lookup
	// for each one, inject the map key into the summary struct.
	for name, summary := range result.ActivityByMode {
		summary.ModeName = name
	}

	for _, platformActivity := range result.ActivityByPlatform {
		for name, summary := range platformActivity.ActivityByMode {
			summary.ModeName = name
		}
	}

	return result, err
}

// GetCurrentMeta will request informatino about the current "meta" in Destiny 2. The parameters
// can be used to filter the meta information returned by the API if a particular activity
// is requested, a set of game modes, or a specific platform.
func (c *Client) GetCurrentMeta(activityHash string, gameModes []string, membershipType string) (*MetaResponse, error) {

	reqURL := c.BaseURL + metaResourcePath
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	vals := url.Values{}
	if activityHash != "" {
		vals.Add("activityHash", activityHash)
	}
	if len(gameModes) != 0 {
		vals.Add("modeType", strings.Join(gameModes, ","))
	}
	if membershipType != "" {
		vals.Add("membershipType", string(membershipType))
	}
	req.URL.RawQuery = vals.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	result := &MetaResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(result)

	return result, err
}
