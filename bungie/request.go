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

func NewCurrentAccountRequest() *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   GetMembershipsForCurrentUserEndpoint,
	}
}

func NewGetCharactersRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{CharactersComponent},
		Body:       nil,
	}
}

func NewUserProfileRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{ProfilesComponent,
			ProfileInventoriesComponent, ProfileCurrenciesComponent, CharactersComponent,
			CharacterInventoriesComponent, CharacterEquipmentComponent, ItemInstancesComponent},
	}
}

func NewGetCurrentEquipmentRequest(membershipType int, membershipID string) *APIRequest {
	return &APIRequest{
		HTTPMethod: "GET",
		Endpoint:   fmt.Sprintf(GetProfileEndpointFormat, membershipType, membershipID),
		Components: []string{CharactersComponent, CharacterEquipmentComponent},
	}
}

func NewPostTransferItemRequest(body map[string]interface{}) *APIRequest {
	return &APIRequest{
		HTTPMethod: "POST",
		Endpoint:   TransferItemEndpointURL,
		Body:       body,
	}
}

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
	}
}
