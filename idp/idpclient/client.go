package idpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

func PrincipalById(authSessionId string, systemBaseUri string, requestedUserId string) (*scim.Principal, error) {
	httpClient := http.Client{}
	userEndpoint, err := userEndpointFor(systemBaseUri, requestedUserId)
	if err != nil {
		return nil, err
	}
	req, nRErr := http.NewRequest("GET", userEndpoint.String(), nil)
	if nRErr != nil {
		return nil, fmt.Errorf("can't create http request for '%v' because: %v", userEndpoint.String(), nRErr)
	}
	req.Header.Set("Authorization", "Bearer "+authSessionId)
	response, doErr := httpClient.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("error calling http GET on '%v' because: %v", userEndpoint.String(), doErr)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		p := scim.Principal{}
		if decErr := json.NewDecoder(response.Body).Decode(&p); decErr != nil {
			return nil, fmt.Errorf("response from Identityprovider '%v' is no valid JSON because: %v", userEndpoint.String(), decErr)
		}
		return &p, nil
	case http.StatusNotFound:
		return nil, errors.New("user not found")
	default:
		responseMsg, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("identityprovider %q returned HTTP-Statuscode '%d' and message %q",
			response.Request.URL, response.StatusCode, responseMsg)
	}
}

func userEndpointFor(systemBaseUri string, userId string) (*url.URL, error) {
	userEndpoint := "/identityprovider/scim/users/" + userId

	userEndpointUrl, vPErr := url.Parse(userEndpoint)
	if vPErr != nil {
		return nil, fmt.Errorf("%v", vPErr)
	}
	base, sBPErr := url.Parse(systemBaseUri)
	if sBPErr != nil {
		return nil, fmt.Errorf("invalid SystemBaseUri '%v' because: %v", systemBaseUri, sBPErr)
	}
	return base.ResolveReference(userEndpointUrl), nil
}
