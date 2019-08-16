package scim_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

const donaldDuckJson = `{"id":"146bc69e-1edf-40f6-bf68-849906998838","userName":"d-velop\\donald","name":{"familyName":"Duck","givenName":"Donald"},"displayName":"Donald Duck","title":"Scrum Duck","emails":[{"value":"donal.duck@entenhausen.de"}],"phoneNumbers":[{"value":"+49 1235 9455-1234"}],"groups":[{"value":"d84b34da-c60e-495e-9a0d-59507630be3a","display":"Developer"},{"value":"759eaed7-4f4e-4fac-a5ef-49f03d0811a1","display":"Scrum People"}],"photos":[{"value":"/identityprovider/scim/photo/donaldbig"}]}`

var donaldDuck = scim.Principal{Id: "146bc69e-1edf-40f6-bf68-849906998838", UserName: "d-velop\\donald", Name: scim.UserName{FamilyName: "Duck", GivenName: "Donald"}, DisplayName: "Donald Duck", Title: "Scrum Duck", Emails: []scim.UserValue{{Value: "donal.duck@entenhausen.de"}}, PhoneNumbers: []scim.UserValue{{Value: "+49 1235 9455-1234"}}, Groups: []scim.UserGroup{{Value: "d84b34da-c60e-495e-9a0d-59507630be3a", Display: "Developer"}, {Value: "759eaed7-4f4e-4fac-a5ef-49f03d0811a1", Display: "Scrum People"}}, Photos: []scim.UserValue{{Value: "/identityprovider/scim/photo/donaldbig"}}}

func TestCanDeserializeSCIMUser(t *testing.T) {
	var u scim.Principal
	err := json.Unmarshal([]byte(donaldDuckJson), &u)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(u, donaldDuck) {
		t.Errorf("Unmarshaled Object wrong: got \n %v want\n %v", u, donaldDuck)
	}
}
