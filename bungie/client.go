package bungie

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

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

func (b *BaseResponse) ErrCode() int      { return b.ErrorCode }
func (b *BaseResponse) ErrStatus() string { return b.ErrorStatus }

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

type BungieNetUser struct {
	MembershipID string `json:"membershipId"`
}

type DestinyMembership struct {
	DisplayName    string `json:"displayName"`
	MembershipType int    `json:"membershipType"`
	MembershipID   string `json:"membershipId"`
}

// GetProfileResponse is the response from the GetProfile endpoint. This data contains information about
// the characeters, inventories, profile inventory, and equipped loadouts.
// https://bungie-net.github.io/multi/operation_get_Destiny2-GetProfile.html#operation_get_Destiny2-GetProfile
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

type ItemListData struct {
	Data *struct {
		Items ItemList `json:"items"`
	} `json:"data"`
}

type CharacterMappedItemListData struct {
	Data map[string]*struct {
		Items ItemList `json:"items"`
	} `json:"data"`
}

// ClientPool is a simple client buffer that will provided round robin access to a collection of Clients.
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
			glg.Errorf("Error executing request: %s", err.Error())
			return err
		}

		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			glg.Warnf("Error executing API request: %s", err.Error())
			return err
		}
		if response.ErrCode() == 36 || response.ErrStatus() == "ThrottleLimitExceededMomentarily" {
			time.Sleep(1 * time.Second)
			retry = true
		}

		glg.Infof("Response for request: %+v", response)
		attempts++
		if retry == false || attempts >= 5 {
			break
		}
	}

	return nil
}
