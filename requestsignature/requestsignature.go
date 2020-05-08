package requestsignature

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
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

type RequestSignatureDto struct {
					EventType string `json:"type"`
					TenantId  string `json:"tenantId"`
					BaseUri   string `json:"baseUri"`
				}

// the DvelopLifeCycleEventPath is path of an app endpoint, that apps must be provide
const DvelopLifeCycleEventPath = "dvelop-cloud-lifecycle-event"

// header who contains relevant headers for signature
const signatureHeaderKey = "x-dv-signature-headers"

// allowed content-type header value
const validContentTypeHeaderValue = "application/json"

// valid time differenz of request
const timeDiff = 5 * time.Minute

// Validate signature of request i.e. for validate HTTP events from cloud center
// The middleware "HandleSignMiddleware" checks the signature of incoming requests. This is important for
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

func HandleSignMiddleware(appSecret []byte, timeNow func() time.Time) func(http.Handler) http.Handler {
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
				http.Error(rw, fmt.Sprintf("wrong life cylce path: got %v", req.URL.Path), http.StatusBadRequest)
				return
			}

			validContentTypeHeaderValue := "application/json"
			if accept := req.Header.Get("content-type"); accept != validContentTypeHeaderValue {
				log.Printf("wrong content-type header found. Got %v want %v", accept, validContentTypeHeaderValue)
				http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			signer := NewRequestSigner(appSecret, timeNow)
			err := signer.ValidateSignedRequest(req)
			if err != nil {
				log.Print("validate signed request failed: ", err)
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
}

func NewRequestSigner(appSecret []byte, timeNow func() time.Time) RequestSigner {
	return &requestSigner{
		appSecret,
		timeNow,
	}
}

// validate signed request as function
func (signer *requestSigner) ValidateSignedRequest(req *http.Request) error {
	if signer.appSecret == nil {
		return errors.New("app secret has not been configured")
	}

	if contentType := req.Header.Get("content-type"); contentType != validContentTypeHeaderValue {
		return errors.New(fmt.Sprintf("wrong accept header found. Got %v want %v", contentType, validContentTypeHeaderValue))
	}

	authorizationHeaderValue, err := signer.isAuthorationHeaderValid(req)
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
	log.Print("normalized request hex hash: ", normalizedRequestHash)

	hmacHexValue := signer.getHmacHash(normalizedRequestHash)
	log.Print("hmac hex hash: ", hmacHexValue)

	if !signer.isAuthorizationHeaderEqualsToCalculatedHmacHex(authorizationHeaderValue, hmacHexValue) {
		return errors.New(fmt.Sprintf("wrong authorization header. Got %v want %v", authorizationHeaderValue, hmacHexValue))
	}
	return nil
}

func (signer *requestSigner) isAuthorizationHeaderEqualsToCalculatedHmacHex(authorizationHeaderValue string, hmacHexValue string) bool {
	return reflect.DeepEqual(authorizationHeaderValue, hmacHexValue)
}

func (signer *requestSigner) isAuthorationHeaderValid(req *http.Request) (string, error) {
	bearerRegex := regexp.MustCompile(`(?m)^(Bearer [[:xdigit:]]+)$`)
	authorizationHeaderValue := req.Header.Get("Authorization")
	if authorizationHeaderValue == "" {
		return "", errors.New("authorization header missing")
	}
	if !bearerRegex.MatchString(authorizationHeaderValue) {
		return "", errors.New(fmt.Sprintf("found authorization header is not a valid Bearer token. Got %v", authorizationHeaderValue))
	}
	authorizationHeaderValue = strings.TrimPrefix(authorizationHeaderValue, "Bearer ")
	return authorizationHeaderValue, nil
}

func (signer *requestSigner) isTimestampValid(req *http.Request) error {
	timestampHeaderValue, err := time.Parse(time.RFC3339, req.Header.Get("x-dv-signature-timestamp"))
	if err != nil {
		return err
	}

	timeNow := signer.now().UTC()

	timeBeforTimestamp := timeNow.Add(-timeDiff)
	timeAfterTimestamp := timeNow.Add(timeDiff)

	if !(timestampHeaderValue.After(timeBeforTimestamp) && timestampHeaderValue.Before(timeAfterTimestamp)) {
		return errors.New(fmt.Sprintf("request is timed out: timestamp from request: %v, current time: %v", timestampHeaderValue.Format(time.RFC3339), timeNow.Format(time.RFC3339)))
	}

	return nil
}

func (signer *requestSigner) getHexHashForNormalizedHeaders(req *http.Request) (hex string, err error) {
	if req.Body == nil {
		return "", errors.New("payload missing")
	}

	body, err := signer.getBodyFromRequest(req)
	if err != nil {
		return "", err
	}

	signedHeaders := strings.Split(req.Header.Get(signatureHeaderKey), ",")
	sort.Strings(signedHeaders)

	normalizedRequest := signer.getNormalizedRequestWithHeaderAndBody(req, signedHeaders, body)

	strNormalizedRequest := strings.Join(normalizedRequest, "\n")
	log.Printf("normalized request: %#v", strNormalizedRequest)

	hashNormalizedRequest := sha256.Sum256([]byte(strNormalizedRequest))

	return strings.ToLower(fmt.Sprintf("%x", hashNormalizedRequest)), nil
}

func (signer *requestSigner) getNormalizedRequestWithHeaderAndBody(req *http.Request, signedHeaders []string, body []byte) []string {
	normalizedHeaders := []string{}

	for _, name := range signedHeaders {
		headerValue := req.Header.Get(name)
		normalizedHeaders = append(normalizedHeaders, fmt.Sprintf("%v:%v", strings.ToLower(name), strings.TrimSpace(headerValue)))
	}

	normalizedRequest := []string{}
	normalizedRequest = append(normalizedRequest, req.Method)
	normalizedRequest = append(normalizedRequest, req.URL.Path)
	normalizedRequest = append(normalizedRequest, req.URL.RawQuery)
	normalizedRequest = append(normalizedRequest, fmt.Sprintf("%v\n", strings.Join(normalizedHeaders, "\n")))
	normalizedRequest = append(normalizedRequest, signer.getHexHashedPayload(body))

	return normalizedRequest
}

func (signer *requestSigner) getBodyFromRequest(req *http.Request) ([]byte, error) {
	if req.GetBody != nil {
		log.Print("get copy of request body")

		var bodyReader io.Reader
		bodyReader, err := req.GetBody()
		if err != nil {
			return nil,  err
		}

		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}

		return  body,nil
	} else {
		log.Print("request.GetBody is nil. Read body and create new request body (reader) with copy of body")

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil,err
		}

		return body, nil
	}
}

func (signer *requestSigner) getHexHashedPayload(payload []byte) string {
	log.Print("payload: ", string(payload))
	hash := sha256.Sum256(payload)
	return strings.ToLower(fmt.Sprintf("%x", hash))
}

func (signer *requestSigner) getHmacHash(normalizedRequestHash string) string {
	hmacHash := hmac.New(sha256.New, signer.appSecret)
	hmacHash.Write([]byte(normalizedRequestHash))
	hmacResult := hmacHash.Sum(nil)
	return strings.ToLower(fmt.Sprintf("%x", hmacResult))
}