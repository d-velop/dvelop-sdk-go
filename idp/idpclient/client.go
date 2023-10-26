// name 'idpclient' is used although it repeats the name of the outer package 'idp'
// in order to improve readability functions like e.g. idpclient.New() or idpclient.HttpClient()
package idpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/d-velop/dvelop-sdk-go/idp/scim"
)

type client struct {
	httpClient     *http.Client
	principalCache Cache
}

// Cache is an interface representing the ability to cache arbitrary items for
// a certain amount of time
type Cache interface {
	// Get an item from the cache. Returns the item or nil, and a bool indicating
	// whether the key was found.
	Get(key string) (item interface{}, found bool)

	// Add an item to the cache, replacing any existing item.
	// If cacheDuration is <= 0 the item never expires
	Set(key string, item interface{}, cacheDuration time.Duration)
}

type Option func(*client) error

// HttpClient explicitly sets the http.Client which should be used to make
// request against the IdentityProvider-App
func HttpClient(h *http.Client) Option {
	return func(c *client) error {
		c.httpClient = h
		return nil
	}
}

func PrincipalCache(pc Cache) Option {
	return func(c *client) error {
		c.principalCache = pc
		return nil
	}
}

// New creates a new Client for the IdentityProvider-App using the following defaults:
//
//   - HttpClient: http.DefaultClient
//   - principalCache: An internal implementation is used
//
// If you don't want to use the defaults provide one or more options to this function.
func New(options ...Option) (*client, error) {
	c := &client{
		httpClient:     http.DefaultClient,
		principalCache: cache.New(cache.DefaultExpiration, 5*time.Minute), // use defaultExpiration to fulfill Set() of Cache interface
	}

	for _, option := range options {
		err := option(c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

var maxAgeRegex = regexp.MustCompile(`(?i)max-age=([^,\s]*)`) // cf. https://regex101.com/

/*
Validate checks if the authSessionId is valid for the tenant specified by systemBaseUri and tenantId.

If the authSessionId is valid, that is it belongs to a principal and has not expired, a none nil *scim.Principal is returned.
Otherwise the returned *scim.Principal is nil.

An error is returned if something unexpected occurred.
The returned error will be of type *url.Error if the remote call to the IdentityProvider-App failed due to a network
connectivity problem or a timout. In case of a timeout the error values Timeout() method will report true.

Use a context with timeout to set a timeout for validate like:

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	p, err := defaultClient.Validate(...)
	if err != nil {
		var urlError *url.Error
		if errors.As (err, &urlError) && urlError.Timeout(){
			// retry request?
		}else{
			return err
		}
	}

ATTENTION: Please note that as successful validation solely asserts, that the user has been successfully identified.
In the vast majority of business cases this is by no means sufficient to grant access to the respective business function.
A proper authorization MUST be implemented to decide if the user should be granted the right to execute the respective function
taking into account the information given in the scim.Principal result.
This is particularly important because the IdentityProvider-App successfully validates external users if configured accordingly.
These external users have been successfully authenticated by an external identity provider like google, facebook,...
(if the admin configured them as external identity provider) but have not been added explicitly to the list of known users of the tenant.
So you don't know much about external users except that they provided valid credentials to one of the configured external
identity providers. For example if google is configured as an external identity provider anyone with a google account
can sign in successfully.
In most of the cases you don't want these users to access your business functions. Exceptions are business cases
in which external users participate such as sharing information with identified but otherwise external users.
Inspect the scim.Principal after a successful validation to distinguish external from internal users
(cf. documentation of scim.Principal for further information).
*/
func (c *client) Validate(ctx context.Context, systemBaseUri string, tenantId string, authSessionId string) (*scim.Principal, error) {
	cacheKey := fmt.Sprintf("%s/%s", tenantId, authSessionId)
	co, found := c.principalCache.Get(cacheKey)
	if found {
		p := co.(scim.Principal)
		return &p, nil
	}

	endpoint := "/identityprovider/validate?allowExternalValidation=true"
	resp, doErr := c.httpGet(ctx, systemBaseUri, authSessionId, endpoint)
	if doErr != nil {
		return nil, fmt.Errorf("error calling http GET on '%s' because: %w", endpoint, doErr)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var p scim.Principal
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return nil, fmt.Errorf("response from Identityprovider '%s' is no valid JSON because: %v", endpoint, err)
		}
		var validFor time.Duration = 0
		cacheControlHeader := resp.Header.Get("Cache-Control")
		matches := maxAgeRegex.FindStringSubmatch(cacheControlHeader)
		if matches != nil {
			d, err := time.ParseDuration(matches[1] + "s")
			if err == nil {
				validFor = d
			}
		}
		if validFor > 0 {
			c.principalCache.Set(cacheKey, p, validFor)
		}
		return &p, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		_, _ = ioutil.ReadAll(resp.Body) // client must read to EOF and close body cf. https://godoc.org/net/http#Client
		return nil, nil
	default:
		responseMsg, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected error. Identityprovider '%s' returned HTTP-Statuscode '%d' and message '%s'",
			resp.Request.URL, resp.StatusCode, responseMsg)
	}
}

/*
GetPrincipalById gets the principal specified by principalId for the tenant specified by systemBaseUri and tenantId.
The authSessionId is used to authorize the request.

If the user exists, a none nil *scim.Principal is returned.
Otherwise the returned *scim.Principal is nil.
*/
func (c *client) GetPrincipalById(ctx context.Context, systemBaseUri string, tenantId string, authSessionId string, principalId string) (*scim.Principal, error) {
	// tenantid not used so far but included to implement a cache without changing the method signature
	endpoint := "/identityprovider/scim/users/" + principalId
	resp, doErr := c.httpGet(ctx, systemBaseUri, authSessionId, endpoint)
	if doErr != nil {
		return nil, fmt.Errorf("error calling http GET on '%s' because: %w", endpoint, doErr)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var p scim.Principal
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return nil, fmt.Errorf("response from Identityprovider '%s' is no valid JSON because: %v", endpoint, err)
		}
		return &p, nil
	case http.StatusForbidden:
		responseMsg, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("user is not allowed to invoke '%s'. Identityprovider returned HTTP-Statuscode '%d' and message '%s'",
			resp.Request.URL, resp.StatusCode, responseMsg)
	case http.StatusNotFound:
		_, _ = ioutil.ReadAll(resp.Body)
		return nil, nil
	default:
		responseMsg, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected error. Identityprovider '%s' returned HTTP-Statuscode '%d' and message '%s'",
			resp.Request.URL, resp.StatusCode, responseMsg)
	}
}

func (c *client) httpGet(ctx context.Context, systemBaseUri string, authSessionId string, absolutePath string) (*http.Response, error) {
	baseUri, baseParseErr := url.Parse(systemBaseUri)
	if baseParseErr != nil {
		return nil, baseParseErr
	}
	resourcePath, _ := url.Parse(absolutePath)
	resourceEndpoint := baseUri.ResolveReference(resourcePath)

	req, nRErr := http.NewRequestWithContext(ctx, http.MethodGet, resourceEndpoint.String(), nil)
	if nRErr != nil {
		return nil, fmt.Errorf("can't create http request for '%s' because: %v", resourceEndpoint, nRErr)
	}
	req.Header.Set("Authorization", "Bearer "+authSessionId)

	return c.httpClient.Do(req)
}
