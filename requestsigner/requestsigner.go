package requestsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RequestSigner interface {
	ValidateSignedRequest(req *http.Request) error
}

// The DvelopLifeCycleEventPath is path of an app endpoint, that apps must be provide
const DvelopLifeCycleEventPath = "dvelop-cloud-lifecycle-event"

// validate signed request as middleware
func HandleSignMiddleware(appSecret []byte, timeDifferenzInMinutes int, timeNow func() time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if appSecret == nil {
				log.Print("error validation signed request because app secret has not been configured")
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			if req.Method != http.MethodPost {
				log.Printf("only POST request can be signed. Got method %v", req.Method)
				http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
			if pathBase := path.Base(req.URL.Path); pathBase != DvelopLifeCycleEventPath {
				log.Printf("path %v is not life cycle path. Life cycle path is %v", req.URL.Path, DvelopLifeCycleEventPath)
				http.Error(rw, "wrong life cylce path", http.StatusBadRequest)
				return
			}
			validAcceptHeaderValue := "application/json"
			if accept := req.Header.Get("accept"); accept != validAcceptHeaderValue {
				log.Printf("wrong accept header found. Got %v want %v", accept, validAcceptHeaderValue)
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			signer := NewRequestSigner(appSecret, timeDifferenzInMinutes, timeNow)
			err := signer.ValidateSignedRequest(req)
			if err != nil {
				log.Print("validate signed request failed: ", err)
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		})
	}
}

type requestSigner struct {
	appSecret             []byte
	timeDifferenzInMinute int
	now                   func() time.Time
}

func NewRequestSigner(appSecret []byte, timeDifferenzInMinute int, timeNow func() time.Time) RequestSigner {
	if timeDifferenzInMinute < 1 {
		timeDifferenzInMinute = 5
	}
	return &requestSigner{
		appSecret,
		timeDifferenzInMinute,
		timeNow,
	}
}

// validate signed request as function
func (signer *requestSigner) ValidateSignedRequest(req *http.Request) error {
	bearerRegex := regexp.MustCompile(`(?m)^(Bearer [[:xdigit:]]+)$`)
	authorizationHeaderValue := req.Header.Get("Authorization")
	if !bearerRegex.MatchString(authorizationHeaderValue) {
		return errors.New(fmt.Sprintf("found authorization header is not a valid Bearer token. Got %v", authorizationHeaderValue))
	}
	authorizationHeaderValue = strings.TrimPrefix(authorizationHeaderValue, "Bearer ")

	validAcceptHeaderValue := "application/json"
	if accept := req.Header.Get("accept"); accept != validAcceptHeaderValue {
		return errors.New(fmt.Sprintf("wrong accept header found. Got %v want %v", accept, validAcceptHeaderValue))
	}

	err := signer.validateTimestamp(req)
	if err != nil {
		return err
	}
	normalizedRequestHash, err := signer.getHexHashForNormalizedHeaders(req)
	if err != nil {
		return err
	}
	hmacHexValue := signer.getHmacHash(normalizedRequestHash)

	if !reflect.DeepEqual(authorizationHeaderValue, hmacHexValue) {
		return errors.New(fmt.Sprintf("wrong authorization header. Got %v want %v", authorizationHeaderValue, hmacHexValue))
	}
	return nil
}

func (signer *requestSigner) validateTimestamp(req *http.Request) error {
	timestampHeaderValue, err := time.Parse(time.RFC3339, req.Header.Get("x-dv-signature-timestamp"))
	if err != nil {
		return err
	}
	diffDuration := time.Duration(signer.timeDifferenzInMinute) * time.Minute
	timeNow := signer.now().UTC()
	timeBeforTimestamp := timeNow.Add(-diffDuration)
	timeAfterTimestamp := timeNow.Add(diffDuration)
	if !(timestampHeaderValue.After(timeBeforTimestamp) && timestampHeaderValue.Before(timeAfterTimestamp)) {
		return errors.New(fmt.Sprintf("request is timed out: timestamp from request: %v | current time before %v minutes: %v | current time in %v minutes: %v",
			timestampHeaderValue.Format(time.RFC3339),
			signer.timeDifferenzInMinute,
			timeBeforTimestamp.Format(time.RFC3339),
			signer.timeDifferenzInMinute,
			timeAfterTimestamp.Format(time.RFC3339),
		))
	}
	return nil
}

func (signer *requestSigner) getHexHashForNormalizedHeaders(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", errors.New("payload missing")
	}
	bodyReader, err := req.GetBody()
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", err
	}

	signedHeaders := strings.Split(req.Header.Get("x-dv-signed-headers"), ",")
	sort.Strings(signedHeaders)
	normalizedHeaders := []string{}
	for _, name := range signedHeaders {
		headerValue := req.Header.Get(name)
		normalizedHeaders = append(normalizedHeaders, fmt.Sprintf("%v:%v", strings.ToLower(name), strings.TrimSpace(headerValue)))
	}
	normalizedRequest := []string{}
	normalizedRequest = append(normalizedRequest, req.Method)
	normalizedRequest = append(normalizedRequest, req.URL.Path)
	normalizedRequest = append(normalizedRequest, req.URL.RawQuery)
	normalizedRequest = append(normalizedRequest, strings.Join(normalizedHeaders, "\n"))
	normalizedRequest = append(normalizedRequest, signer.getHexHashedPayload(body))

	strNormalizedRequest := strings.Join(normalizedRequest, "\n")
	hashNormalizedRequest := sha256.Sum256([]byte(strNormalizedRequest))
	return strings.ToLower(fmt.Sprintf("%x", hashNormalizedRequest)), nil
}

func (signer *requestSigner) getHexHashedPayload(payload []byte) string {
	hash := sha256.Sum256(payload)
	return strings.ToLower(fmt.Sprintf("%x", hash))
}

func (signer *requestSigner) getHmacHash(normalizedRequestHash string) string {
	hmacHash := hmac.New(sha256.New, signer.appSecret)
	hmacHash.Write([]byte(normalizedRequestHash))
	hmacResult := hmacHash.Sum(nil)
	return strings.ToLower(fmt.Sprintf("%x", hmacResult))
}
