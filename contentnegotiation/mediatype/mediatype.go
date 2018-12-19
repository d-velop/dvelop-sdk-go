// Package mediatype provides HTTP/1.1 compliant content negotiation for mediatypes
//
// cf. https://tools.ietf.org/html/rfc7231#section-5.3.2
package mediatype

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrNotSupported     = errors.New("mediatype: not supported")
	mediarangeRegexp, _ = regexp.Compile("([^/]+)/([^/]+)")
)

// Mediatype represents a HTTP/1.1 compliant mediatyp with Maintype and Subtype
type Mediatype struct {
	Maintype string
	Subtype  string
}

// Negotiate negotiates the mediatype based on the acceptHeader from the http request
// and the supportedTypes.
//
// It returns the negotiated mediatype or an error if none of the requested mediatypes is supported.
//
// Example:
//	func handler(w http.ResponseWriter, req *http.Request) {
//		negotiatedType, err := mediatype.Negotiate(req.Header.Get("Accept"), []string{"text/html","application/json"})
//		if err != nil {
//			http.Error(w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
//			return
//		}
//		switch negotiatedType.String() {
//		case "text/html":
//			w.Header().Set("content-type", negotiatedType.String()+";charset=utf-8")
//			fmt.Fprint(w,"<!DOCTYPE html><html><head></head><body><h1>Hello</h1></body></html>")
//		case "application/json":
//			w.Header().Set("content-type", negotiatedType.String()+";charset=utf-8")
//			json.NewEncoder(w).Encode(&Dto{Title:"title",Version:"1.2.3"})
//		}
//	}
func Negotiate(acceptHeader string, supportedTypes []string) (*Mediatype, error) {
	if len(supportedTypes) == 0 {
		return nil, ErrNotSupported
	}

	if acceptHeader == "" {
		return parseMediatype(supportedTypes[0]), nil
	}

	tokens := strings.Split(acceptHeader, ",")
	mediaranges := make([]*headervalue, len(tokens))
	for k, r := range tokens {
		x, _ := parseHeaderValue(r)
		mediaranges[k] = x
	}
	sort.Sort(sort.Reverse(headervalues(mediaranges)))

	for _, mr := range mediaranges {
		for _, t := range supportedTypes {
			mt := parseMediatype(t)
			if mt.matches(mr.Value) {
				return mt, nil
			}
		}
	}

	return nil, ErrNotSupported
}

func parseMediatype(s string) *Mediatype {
	mt := mediarangeRegexp.FindStringSubmatch(s)
	if len(mt) < 3 {
		return &Mediatype{Maintype: "", Subtype: ""}
	}
	return &Mediatype{Maintype: mt[1], Subtype: mt[2]}
}

func (mediatype *Mediatype) matches(mediaRange string) bool {
	mr := parseMediatype(mediaRange)
	return (mediatype.Maintype == mr.Maintype || mr.Maintype == "*") &&
		(mediatype.Subtype == mr.Subtype || mr.Subtype == "*")
}

func (mediatype *Mediatype) String() string {
	return fmt.Sprintf("%v/%v", mediatype.Maintype, mediatype.Subtype)
}
