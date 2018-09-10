package bungie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/getsentry/raven-go"

	"github.com/kpango/glg"
)

// StatusResponse is used as the generic response parameter for the deserialized response
// from the generic Client Execute calls. One of the below structs should be used as the
// concrete type for the request's response to be deserialized into.
type StatusResponse interface {
	ErrCode() int
	ErrStatus() string
}

// BaseResponse represents the data returned as part of all of the Bungie API
// requests.
type BaseResponse struct {
	ErrorCode       int         `json:"ErrorCode"`
	ThrottleSeconds int         `json:"ThrottleSeconds"`
	ErrorStatus     string      `json:"ErrorStatus"`
	Message         string      `json:"Message"`
	MessageData     interface{} `json:"MessageData"`
}

// ErrCode returns the err code field from a Bungie response
func (b *BaseResponse) ErrCode() int { return b.ErrorCode }

// ErrStatus returns the status string provided in the Bungie response
func (b *BaseResponse) ErrStatus() string { return b.ErrorStatus }
func (b *BaseResponse) String() string {
	if b != nil {
		return fmt.Sprintf("%+v", *b)
	}
	return "<nil>"
}

// CurrentUserMembershipsResponse contains information about the membership data for the currently
// authorized user. The request for this information will use the access_token to determine
// the current user
// https://bungie-net.github.io/multi/operation_get_User-GetMembershipDataForCurrentUser.html#operation_get_User-GetMembershipDataForCurrentUser
type CurrentUserMembershipsResponse struct {
	*BaseResponse
	Response *struct {
		DestinyMemberships []*DestinyMembership `json:"destinyMemberships"`
		BungieNetUser      *BungieNetUser       `json:"bungieNetUser"`
	} `json:"Response"`
}

// CurrentUserMemberships will hold the current user's Bungie.net membership data
// as well as the Destiny membership data for their most recently played character.
type CurrentUserMemberships struct {
	BungieNetUser     *BungieNetUser
	DestinyMembership *DestinyMembership
}

// BungieNetUser holds fields relating to a specific Bungie membership
type BungieNetUser struct {
	MembershipID string `json:"membershipId"`
}

// DestinyMembership holds information about a specific Destiny membership
type DestinyMembership struct {
	DisplayName    string `json:"displayName"`
	MembershipType int    `json:"membershipType"`
	MembershipID   string `json:"membershipId"`
}

// CharacterProgressionResponse is the JSON response representation of the character progression
// data from the GetProfile endpoint.
type CharacterProgressionResponse struct {
	*BaseResponse
	Response *struct {
		CharacterProgressions *struct {
			Data map[string]*CharacterProgression `json:"data"`
		} `json:"characterProgressions"`
	} `json:"Response"`
}

func (r *CharacterProgressionResponse) valorProgressionForChar(charID string) *DestinyProgression {

	charProgress := r.Response.CharacterProgressions.Data[charID]
	if charProgress == nil {
		return nil
	}

	return charProgress.Progressions[valorHash]
}

func (r *CharacterProgressionResponse) valorProgression() *DestinyProgression {
	return r.progression(valorHash)
}

func (r *CharacterProgressionResponse) gloryProgressionForChar(charID string) *DestinyProgression {

	charProgress := r.Response.CharacterProgressions.Data[charID]
	if charProgress == nil {
		return nil
	}

	return charProgress.Progressions[gloryHash]
}

func (r *CharacterProgressionResponse) gloryProgression() *DestinyProgression {
	return r.progression(gloryHash)
}

func (r *CharacterProgressionResponse) infamyProgression() *DestinyProgression {
	return r.progression(infamyHash)
}

func (r *CharacterProgressionResponse) progression(hash string) *DestinyProgression {

	charProgress := r.Response.CharacterProgressions.Data
	if charProgress == nil {
		return nil
	}

	for _, progress := range charProgress {
		return progress.Progressions[infamyHash]
	}

	return nil
}

// CharacterProgression contains data for different progressions tied to a specific character
type CharacterProgression struct {
	Progressions map[string]*DestinyProgression   `json:"progressions"`
	Factions     map[string]*FactionProgression   `json:"factions"`
	Milestones   map[string]*MilestoneProgression `json:"milestones"`
	// NOTE: Not using these two yet, not sure what they could be used for
	// Quests
	// uninstancedItemObjectives
}

// BaseProgression contains data relevant to all of the different progression types
type BaseProgression struct {
	ProgressionHash     int `json:"progressionHash"`
	DailyProgress       int `json:"dailyProgress"`
	DailyLimit          int `json:"dailyLimit"`
	WeeklyProgress      int `json:"weeklyProgress"`
	WeeklyLimit         int `json:"weeklyLimit"`
	CurrentProgress     int `json:"currentProgress"`
	Level               int `json:"level"`
	LevelCap            int `json:"levelCap"`
	StepIndex           int `json:"stepIndex"`
	ProgressToNextLevel int `json:"progressToNextLevel"`
	NextLevelAt         int `json:"nextLevelAt"`
}

func (b *BaseProgression) String() string {
	return fmt.Sprintf("%+v", *b)
}

// DestinyProgression contains data about progression through different Destiny related achievements
type DestinyProgression struct {
	*BaseProgression
}

func (p *DestinyProgression) String() string {
	return fmt.Sprintf("%+v", *p)
}

// FactionProgression wraps data related to the progression through levels related
// to the different factions
type FactionProgression struct {
	*BaseProgression
	FactionHash        int `json:"factionHash"`
	FactionVendorIndex int `json:"factionVendorIndex"`
}

func (p *FactionProgression) String() string {
	return fmt.Sprintf("%+v", *p)
}

// MilestoneProgression contains data about progress through different milestones
type MilestoneProgression struct {
	MilestoneHash int `json:"milestoneHash"`
	// NOTE: Not sure how much information from here could be provided. Seems
	// like it could be more information than could be useful
	// AvailableQuests []
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// GetProfileResponse is the response from the GetProfile endpoint. This data contains
// information about the characeters, inventories, profile inventory, and equipped loadouts.
//https://bungie-net.github.io/multi/operation_get_Destiny2-GetProfile.html#operation_get_Destiny2-GetProfile
type GetProfileResponse struct {
	*BaseResponse
	Response *struct {
		CharacterInventories *CharacterMappedItemListData `json:"characterInventories"`
		CharacterEquipment   *CharacterMappedItemListData `json:"characterEquipment"`
		ProfileInventory     *ItemListData                `json:"profileInventory"`
		ProfileCurrencies    *ItemListData                `json:"profileCurrencies"`
		ItemComponents       *struct {
			Instances *struct {
				Data map[string]*ItemInstance `json:"data"`
			} `json:"instances"`
		} `json:"itemComponents"`
		Profile *struct {
			//https://bungie-net.github.io/multi/schema_Destiny-Entities-Profiles-DestinyProfileComponent.html#schema_Destiny-Entities-Profiles-DestinyProfileComponent
			Data *struct {
				UserInfo *struct {
					MembershipType int    `json:"membershipType"`
					MembershipID   string `json:"membershipId"`
					DisplayName    string `json:"displayName"`
				} `json:"userInfo"`
			} `json:"data"`
		} `json:"profile"`
		Characters *struct {
			Data CharacterMap `json:"data"`
		} `json:"Characters"`
	} `json:"Response"`
}

func (r *GetProfileResponse) membershipID() string {
	if r.Response.Profile == nil {
		return ""
	}

	return r.Response.Profile.Data.UserInfo.MembershipID
}

func (r *GetProfileResponse) membershipType() int {
	if r.Response.Profile == nil {
		return 0
	}

	return r.Response.Profile.Data.UserInfo.MembershipType
}

func (r *GetProfileResponse) character(charID string) *Character {
	if r.Response.Characters == nil {
		return nil
	}

	return r.Response.Characters.Data[charID]
}

func (r *GetProfileResponse) instanceData(ID string) *ItemInstance {
	if ID == "" || r.Response.ItemComponents == nil ||
		r.Response.ItemComponents.Instances == nil {
		return nil
	}

	return r.Response.ItemComponents.Instances.Data[ID]
}

// ItemListData contains the list of Items in the format returned by the Bungie.net API
type ItemListData struct {
	Data *struct {
		Items ItemList `json:"items"`
	} `json:"data"`
}

// CharacterMappedItemListData contains the lists of item data mapped by the character ID
// to which they are associated.
type CharacterMappedItemListData struct {
	Data map[string]*struct {
		Items ItemList `json:"items"`
	} `json:"data"`
}

// ClientPool is a simple client buffer that will provided round robin access to a collection
// of Clients.
type ClientPool struct {
	Clients []*Client
	current int
}

// NewClientPool is a convenience initializer to create a new collection of Clients.
func NewClientPool() *ClientPool {

	addresses := readClientAddresses()
	clients := make([]*Client, 0, len(addresses))
	for _, addr := range addresses {
		client, err := NewCustomAddrClient(addr)
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Error creating custom ipv6 client: %s", err.Error())
			continue
		}

		clients = append(clients, client)
	}
	if len(clients) == 0 {
		clients = append(clients, &Client{Client: http.DefaultClient})
	}

	return &ClientPool{
		Clients: clients,
	}
}

// Get will return a pointer to the next Client that should be used.
func (pool *ClientPool) Get() *Client {
	c := pool.Clients[pool.current]
	if pool.current == (len(pool.Clients) - 1) {
		pool.current = 0
	} else {
		pool.current++
	}

	return c
}

func readClientAddresses() (result []string) {
	result = make([]string, 0, 32)

	in, err := os.OpenFile("local_clients.txt", os.O_RDONLY, 0644)
	if err != nil {
		glg.Warn("Local clients list does not exist, using the default...")
		return
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		addr := scanner.Text()
		if addr != "" {
			result = append(result, addr)
		}
	}

	if err != nil {
		raven.CaptureError(err, nil)
		glg.Errorf("Failed to read local clients: %s", err.Error())
	}

	return
}

// Client is a type that contains all information needed to make requests to the
// Bungie API.
type Client struct {
	*http.Client
	Address     string
	AccessToken string
	APIToken    string
}

// NewCustomAddrClient will create a new Bungie Client instance with the provided local IP address.
func NewCustomAddrClient(address string) (*Client, error) {

	localAddr, err := net.ResolveIPAddr("ip6", address)
	if err != nil {
		return nil, err
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
	}

	httpClient := &http.Client{Transport: transport}

	return &Client{Client: httpClient, Address: address}, nil
}

// AddAuthValues will add the specified access token and api key to the provided client
func (c *Client) AddAuthValues(accessToken, apiKey string) {
	c.APIToken = apiKey
	c.AccessToken = accessToken
}

// AddAuthHeadersToRequest will handle adding the authentication headers from the
// current client to the specified Request.
func (c *Client) AddAuthHeadersToRequest(req *http.Request) {
	authHeaders := map[string]string{
		"X-Api-Key":     c.APIToken,
		"Authorization": "Bearer " + c.AccessToken,
	}
	for key, val := range authHeaders {
		req.Header.Add(key, val)
	}
}

// GetCurrentAccount will request the user info for the current user
// based on the OAuth token provided as part of the request.
func (c *Client) GetCurrentAccount() (*CurrentUserMemberships, error) {

	accountResponse := CurrentUserMembershipsResponse{}
	c.Execute(NewCurrentAccountRequest(), &accountResponse)

	glg.Debugf("Found %d Destiny memberships", len(accountResponse.Response.DestinyMemberships))

	// If the user only has a single destiny membership, just use that then
	if len(accountResponse.Response.DestinyMemberships) == 1 {
		return &CurrentUserMemberships{
			BungieNetUser:     accountResponse.Response.BungieNetUser,
			DestinyMembership: accountResponse.Response.DestinyMemberships[0],
		}, nil
	}

	allChars := make(CharacterList, 0, 9)
	destinyMembershipLookup := make(map[string]*DestinyMembership)
	for _, destinyMembership := range accountResponse.Response.DestinyMemberships {
		destinyMembershipLookup[destinyMembership.MembershipID] = destinyMembership

		allChars = append(allChars,
			c.getCharacters(NewGetCharactersRequest(destinyMembership.MembershipType, destinyMembership.MembershipID))...)
	}

	latestDestinyMembership := accountResponse.Response.DestinyMemberships[0]
	sort.Sort(sort.Reverse(LastPlayedSort(allChars)))
	glg.Debugf("Found all membership characters: %+v", allChars)
	if len(allChars) > 0 {
		latestDestinyMembership = destinyMembershipLookup[allChars[0].MembershipID]
	}

	return &CurrentUserMemberships{
		BungieNetUser:     accountResponse.Response.BungieNetUser,
		DestinyMembership: latestDestinyMembership,
	}, nil
}

func (c *Client) getCharacters(request *APIRequest) CharacterList {

	profile := &GetProfileResponse{}
	c.Execute(request, profile)

	chars := make(CharacterList, 0, 3)
	if profile.Response == nil || profile.Response.Characters == nil {
		return chars
	}

	for _, char := range profile.Response.Characters.Data {
		chars = append(chars, char)
	}

	return chars
}

// Execute is a generic request execution method that will send the passed request
// on to the Bungie API using the configured client. The response is then deserialized into
// the response object provided.
func (c *Client) Execute(request *APIRequest, response StatusResponse) error {

	glg.Debugf("Client local address: %s", c.Address)

	// TODO: This retry logic should probably be added to a middleware type function
	retry := true
	attempts := 0
	var resp *http.Response
	var err error

	for {
		retry = false
		var req *http.Request

		glg.Warnf("Executing request: %+v", request)
		if request.Body != nil && (len(request.Body) > 0) {
			glg.Warn("Setting the requests body...")
			jsonBody, _ := json.Marshal(request.Body)
			bodyReader := strings.NewReader(string(jsonBody))

			req, _ = http.NewRequest(request.HTTPMethod, request.Endpoint, bodyReader)
		} else {
			req, _ = http.NewRequest(request.HTTPMethod, request.Endpoint, nil)
		}

		req.Header.Add("Content-Type", "application/json")
		c.AddAuthHeadersToRequest(req)

		if request.Components != nil && (len(request.Components) > 0) {
			vals := url.Values{}
			vals.Add("components", strings.Join(request.Components, ","))
			req.URL.RawQuery = vals.Encode()
		}

		resp, err = c.Do(req)
		if err != nil {
			raven.CaptureError(err, nil)
			glg.Errorf("Error executing request: %s", err.Error())
			return err
		}

		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			raven.CaptureError(err, nil)
			glg.Warnf("Error executing API request: %s", err.Error())
			return err
		}
		if response.ErrCode() == 36 || response.ErrStatus() == "ThrottleLimitExceededMomentarily" {
			time.Sleep(1 * time.Second)
			retry = true
		}

		glg.Successf("Response for request: %+v", response)
		attempts++
		if retry == false || attempts >= 5 {
			break
		}
	}

	return nil
}
