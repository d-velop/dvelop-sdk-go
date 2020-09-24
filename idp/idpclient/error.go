package idpclient

import (
	"fmt"
	"net/url"
)

type IdpClientError struct {
	ErrorMessage string
	RequestURL   url.URL
	StatusCode   int
	ResponseMsg  string
}

func NewIdpClientError(errorMessage string, requestUrl url.URL, statusCode int, responseMsg string) *IdpClientError {
	return &IdpClientError{
		ErrorMessage: errorMessage,
		RequestURL:   requestUrl,
		StatusCode:   statusCode,
		ResponseMsg:  responseMsg,
	}
}

func (idpClientError IdpClientError) Error() string {
	return fmt.Sprintf("%s Identityprovider '%v' returned HTTP-Statuscode '%d' and message '%s'", idpClientError.ErrorMessage, idpClientError.RequestURL, idpClientError.StatusCode, idpClientError.ResponseMsg)
}
