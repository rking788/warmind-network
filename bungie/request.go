package bungie

import (
	"fmt"
)

// APIRequest is a generic request object that can be sent to a bungie.Client and
// the client will automatically handle setting up the request body, url parameters,
// and full url (including endpoint).
type APIRequest struct {
	HTTPMethod string
	Endpoint   string
	Components []string
	Body       map[string]interface{}
}

// NewCurrentAccountRequest is a helper function for creating a request object to get the
// memberships for a specific user.
func NewCurrentAccountRequest() *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   GetMembershipsForCurrentUserEndpoint,
	}
}

// NewGetCharactersRequest is a helper function for getting the characters for a specific user
func NewGetCharactersRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{CharactersComponent},
		Body:       nil,
	}
}

// NewUserProfileRequest is a helper function for creating a new request that can be used
// to load all profile data for a specific user on the specified platform and given membership.
func NewUserProfileRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{ProfilesComponent,
			ProfileInventoriesComponent, ProfileCurrenciesComponent, CharactersComponent,
			CharacterInventoriesComponent, CharacterEquipmentComponent, ItemInstancesComponent},
	}
}

// NewGetCurrentEquipmentRequest is a helper function for creating a request that can be used
// to load the equipment for all characters for the given membership and platform.
func NewGetCurrentEquipmentRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{CharactersComponent, CharacterEquipmentComponent},
	}
}

// NewGetProgressionsRequest will be an APIRequest initialized with the correct values
// to request the character progression data from the Bungie.net API.
func NewGetProgressionsRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{CharacterProgressionsComponent},
	}
}

// NewPostTransferItemRequest is a helper function for creating a new request to send an
// API request to transfer one or more items.
func NewPostTransferItemRequest(body map[string]interface{}) *APIRequest {
	return &APIRequest{
		HTTPMethod: "POST",
		Endpoint:   TransferItemEndpointURL,
		Body:       body,
	}
}

// NewPostEquipItem is a helper function for creating a new request to send an
// API request to equip an item.
func NewPostEquipItem(body map[string]interface{}, isMultipleItems bool) *APIRequest {
	var endpoint string
	if isMultipleItems {
		endpoint = EquipMultiItemsEndpointURL
	} else {
		endpoint = EquipSingleItemEndpointURL
	}

	return &APIRequest{
		HTTPMethod: "POST",
		Endpoint:   endpoint,
		Body:       body,
	}
}
