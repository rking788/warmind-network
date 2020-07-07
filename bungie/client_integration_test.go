// +build integration

package bungie

import (
	"testing"
)

const (
	apiKey      = "<REAL_API_KEY>"
	accessToken = "<REAL_ACCESS_TOKEN>"
)

func TestGetProfileForCurrentUser(t *testing.T) {

	clients := NewClientPool()
	c := clients.Get()
	c.AddAuthValues(accessToken, apiKey)
	profile, err := GetProfileForCurrentUser(c, true)
	if err != nil {
		t.Fatalf("Error loading profile for current user: %+v\n", err.Error())
	}

	if profile == nil {
		t.Fatal("Nil profile found for current user")
	} else if profile.MembershipID != "4611686018437694484" && profile.MembershipType != 1 {
		t.Fatalf("Incorrect Destiny membership loaded for the current user: Expected(%s[%d]) Actual(%s[%d])", "4611686018437694484", 1, profile.MembershipID, profile.MembershipType)
	} else if len(profile.Characters) != 3 {
		t.Fatalf("Incorrect number of characters found on profile: Expected(3) Actual(%d)", len(profile.Characters))
	}
}
