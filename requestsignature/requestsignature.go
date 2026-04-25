package requestsignature

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
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

type loggerFunc func(ctx context.Context, message string)

type Dto struct {
	EventType string `json:"type"`
	TenantId  string `json:"tenantId"`
	BaseUri   string `json:"baseUri"`
}

// the DvelopLifeCycleEventPath is path of an app endpoint, that apps must be provide
const DvelopLifeCycleEventPath = "dvelop-cloud-lifecycle-event"

// header who contains relevant headers for signature
const signatureHeaderKey = "x-dv-signature-headers"

// header who contains timestamp of request
const timestampHeaderKey = "x-dv-signature-timestamp"

// allowed content-type header value
const validContentTypeHeaderValue = "application/json"

// valid time differenz of request
const timeDiff = 5 * time.Minute

// Validate signature of request i.e. for validate HTTP events from cloud center
// The middleware "HandleCloudSignatureMiddleware" checks the signature of incoming requests. This is important for
// cloud center to app authentication. The cloudcenter make an POST request to app with a signature. The middleware
// checks if request is a POST request and the content-type header is set to "application/json".
// If the requested signature is valid, then your handler is invoke to handle the request. If the
// signature is invalid, the middleware returns the HTTP error 403 "Forbidden" and log the reason to your application log.
//
// The parameter for the "appSecret" is the base64 decoded app secret string of your app as byte array.
//
// More information about signature algorithm please visit the following documentation:
// 	https://developer.d-velop.de/documentation/ccapi/en/cloud-center-api-199623589.html
//
// Example:
//	func main() {
//		// replace `Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=` with your app secret (base64-string)
//		myAppSecret, err := base64.StdEncoding.DecodeString(`Rg9iJXX0Jkun9u4Rp6no8HTNEdHlfX9aZYbFJ9b6YdQ=`)
//		if err != nil {
//			panic(err)
//		}
//		mux := http.NewServeMux()
//		// the path must a ressource for dvelop-cloud-lifecycle-event
//		path := "/app/dvelop-cloud-lifecycle-event"
//		mux.Handle(path, requestsignatur.HandleSignaturValidation(myAppSecret, time.Now)(eventHandler()))
//	}
//
//	func eventHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
//			eventDto := new(requestsignature.RequestSignatureDto)
//			err := json.NewDecoder(req.Body).Decode(eventDto)
//			if err != nil {
//				log.Print(err)
//				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//				return
//			}
//			doSomeStuff(eventDto)
//		})
//	}

func HandleCloudSignatureMiddleware(appSecret []byte, timeNow func() time.Time, logError, logInfo loggerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			if appSecret == nil {
				logError(ctx, "validation signed request failed because app secret has not been configured")
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if req.Method != http.MethodPost {
				logError(ctx, fmt.Sprintf("only POST request can be signed. Got method %v", req.Method))
				http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			if pathBase := path.Base(req.URL.Path); pathBase != DvelopLifeCycleEventPath {
				logError(ctx, fmt.Sprintf("path %v is not life cycle path. Life cycle path must %v", req.URL.Path, DvelopLifeCycleEventPath))
				http.Error(rw, fmt.Sprintf("wrong life cylce path: got %v", req.URL.Path), http.StatusBadRequest)
				return
			}

			if accept := req.Header.Get("content-type"); accept != validContentTypeHeaderValue {
				logError(ctx, fmt.Sprintf("wrong content-type header found. Got %v want %v", accept, validContentTypeHeaderValue))
				http.Error(rw, fmt.Sprintf("%s: please use content-type '%s'", http.StatusText(http.StatusBadRequest), validContentTypeHeaderValue), http.StatusNotAcceptable)
				return
			}

			signer := NewRequestSigner(appSecret, timeNow, logInfo)
			err := signer.ValidateSignedRequest(req)
			if err != nil {
				logError(ctx, fmt.Sprintf("validate signed request failed: %v", err))
				http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}

// Validate signature of request as function i.e. for validate HTTP events from cloud center
// The function "ValidateSignedRequest" in "RequestSignatureValidator" checks the signature of a requests.
// This is important for cloud center to app authentication. The cloudcenter make an POST request to app with a signature.
// It checks if request is a POST request and the content-type header is set to "application/json". Then an own signature
// will be calculated by information from header "dv-signature-headers" and a hash of request body. If the calculcated
// signature is equals to signature of Authorization-header, the signature in request is valid. If signature is valid,
// no error is returned from the function. Otherwise it returns an error and you must abort the request by returning
// HTTP error 403 "Forbidden".
//
// The parameter for the "appSecret" is the base64 decoded app secret string of your app as byte array.
//
// More information about signature algorithm please visit the following documentation on
// 	https://developer.d-velop.de/documentation/ccapi/en/cloud-center-api-199623589.html
//
// Example:
//	func eventHandler() http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
//			// replace `Zm9vYmFy` with your app secret (base64-string)
//			myAppSecret, err := base64.StdEncoding.DecodeString(`Zm9vYmFy`)
//			if err != nil {
//				panic(err)
//			}
//			signatureValidator := NewRequestSignaturValidator(myAppSecret, time.Now)
//			err = signatureValidator.ValidateSignedRequest(req)
//			if err != nil {
//				log.Print(err)
//				http.Error(w, err.Error(), http.StatusInternalServerError)
//				return
//			}
//
//			eventDto := &struct {
//				EventType string `json:"type"`
//				TenantId  string `json:"tenantId"`
//				BaseUri   string `json:"baseUri"`
//			}{}
//			err = json.NewDecoder(req.Body).Decode(eventDto)
//			if err != nil {
//				log.Print(err)
//				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//				return
//			}
//			doSomeStuff(eventDto)
//	})
//}

type requestSigner struct {
	appSecret []byte
	now       func() time.Time
	logInfo   loggerFunc
}

func NewRequestSigner(appSecret []byte, timeNow func() time.Time, logInfo loggerFunc) RequestSigner {
	return &requestSigner{
		appSecret,
		timeNow,
		logInfo,
	}
}

// validate signed request as function
func (signer *requestSigner) ValidateSignedRequest(req *http.Request) error {
	if len(signer.appSecret) == 0 {
		return fmt.Errorf("app secret has not been configured")
	}

	if contentType := req.Header.Get("content-type"); contentType != validContentTypeHeaderValue {
		return fmt.Errorf("wrong accept header found. Got %v want %v", contentType, validContentTypeHeaderValue)
	}

	authorizationHeaderValue, err := signer.isAuthorizationHeaderValid(req)
	if err != nil {
		return err
	}

	err = signer.isTimestampValid(req)
	if err != nil {
		return err
	}

	normalizedRequestHash, err := signer.getHexHashForNormalizedHeaders(req)
	if err != nil {
		return err
	}

	hmacHexValue := signer.getHmacHash(normalizedRequestHash)

	if !signer.isAuthorizationHeaderEqualsToCalculatedHmacHex(authorizationHeaderValue, hmacHexValue) {
		return fmt.Errorf("wrong signature in authorization header. Got %v want %v", authorizationHeaderValue, hmacHexValue)
	}

	signer.logInfo(req.Context(), "received signature is valid")

	return nil
}

func (signer *requestSigner) isAuthorizationHeaderEqualsToCalculatedHmacHex(authorizationHeaderValue string, hmacHexValue string) bool {
	return reflect.DeepEqual(authorizationHeaderValue, hmacHexValue)
}

func (signer *requestSigner) isAuthorizationHeaderValid(req *http.Request) (string, error) {
	bearerRegex := regexp.MustCompile(`(?m)^(Bearer [[:xdigit:]]+)$`)
	authorizationHeaderValue := req.Header.Get("Authorization")
	if authorizationHeaderValue == "" {
		return "", fmt.Errorf("authorization header missing")
	}
	if !bearerRegex.MatchString(authorizationHeaderValue) {
		return "", fmt.Errorf("found authorization header is not a valid Bearer token. Got %v", authorizationHeaderValue)
	}
	authorizationHeaderValue = strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	return authorizationHeaderValue, nil
}

func (signer *requestSigner) isTimestampValid(req *http.Request) error {
	signer.logInfo(req.Context(), "validate timestamp from request header")
	timestampHeaderValue, err := time.Parse(time.RFC3339, req.Header.Get(timestampHeaderKey))
	if err != nil {
		return err
	}

	timeNow := signer.now().UTC()

	timeBeforTimestamp := timeNow.Add(-timeDiff)
	timeAfterTimestamp := timeNow.Add(timeDiff)

	if !(timestampHeaderValue.After(timeBeforTimestamp) && timestampHeaderValue.Before(timeAfterTimestamp)) {
		return fmt.Errorf("request is timed out: timestamp from request: %v, current time: %v", timestampHeaderValue.Format(time.RFC3339), timeNow.Format(time.RFC3339))
	}

	return nil
}

func (signer *requestSigner) getHexHashForNormalizedHeaders(req *http.Request) (hex string, err error) {
	if req.Body == nil {
		return "", fmt.Errorf("payload missing")
	}

	body, err := signer.getBodyFromRequest(req)
	if err != nil {
		return "", err
	}

	signedHeaders := strings.Split(req.Header.Get(signatureHeaderKey), ",")
	sort.Strings(signedHeaders)

	normalizedRequest := signer.getNormalizedRequestWithHeaderAndBody(req, signedHeaders, body)

	strNormalizedRequest := strings.Join(normalizedRequest, "\n")
	signer.logInfo(req.Context(), fmt.Sprintf("normalized request: %#v", strNormalizedRequest))

	hashNormalizedRequest := sha256.Sum256([]byte(strNormalizedRequest))

	signer.logInfo(req.Context(), "hashing normalized request")

	return strings.ToLower(fmt.Sprintf("%x", hashNormalizedRequest)), nil
}

func (signer *requestSigner) getNormalizedRequestWithHeaderAndBody(req *http.Request, signedHeaders []string, body []byte) []string {
	var normalizedHeaders []string

	for _, name := range signedHeaders {
		headerValue := req.Header.Get(name)
		normalizedHeaders = append(normalizedHeaders, fmt.Sprintf("%v:%v", strings.ToLower(name), strings.TrimSpace(headerValue)))
	}

	var normalizedRequest []string
	normalizedRequest = append(normalizedRequest, req.Method)
	normalizedRequest = append(normalizedRequest, req.URL.Path)
	normalizedRequest = append(normalizedRequest, req.URL.RawQuery)
	normalizedRequest = append(normalizedRequest, fmt.Sprintf("%v\n", strings.Join(normalizedHeaders, "\n")))

	signer.logInfo(req.Context(), "hashing body")
	normalizedRequest = append(normalizedRequest, signer.getHexHashedPayload(body))

	return normalizedRequest
}

func (signer *requestSigner) getBodyFromRequest(req *http.Request) ([]byte, error) {
	if req.GetBody != nil {
		signer.logInfo(req.Context(), "get a copy of request body")

		bodyReader, err := req.GetBody()
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	signer.logInfo(req.Context(), "request.GetBody is nil. Read body and create new request body with read body data")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return body, nil
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
