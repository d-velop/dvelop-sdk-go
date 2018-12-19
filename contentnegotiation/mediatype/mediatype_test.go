package mediatype_test

import (
	"github.com/d-velop/dvelop-sdk-go/contentnegotiation/mediatype"
	"testing"
)

func TestRequestedTypeNotSupported_ReturnsErrorNotSupported(t *testing.T) {
	(&negotiateWith{acceptHeader: "text/html", supportedTypes: []string{}}).shouldReturnErrorNotSupported(t)
	(&negotiateWith{acceptHeader: "application/json", supportedTypes: []string{}}).shouldReturnErrorNotSupported(t)
	(&negotiateWith{acceptHeader: "application/json", supportedTypes: []string{"text/html"}}).shouldReturnErrorNotSupported(t)
	(&negotiateWith{acceptHeader: "text/*", supportedTypes: []string{"application/json"}}).shouldReturnErrorNotSupported(t)
}

func TestEmptyAcceptHeaderAndSupportedTypesIsEmpty_ReturnsErrorNotSupported(t *testing.T) {
	(&negotiateWith{acceptHeader: "", supportedTypes: []string{}}).shouldReturnErrorNotSupported(t)
}

func TestInvalidAcceptHeader_ReturnsErrorNotSupported(t *testing.T) {
	(&negotiateWith{acceptHeader: "*/", supportedTypes: []string{"text/html"}}).shouldReturnErrorNotSupported(t)
}

func TestEmptyAcceptHeaderAndSupportedTypesHasOneElement_ReturnsSupportedType(t *testing.T) {
	(&negotiateWith{acceptHeader: "", supportedTypes: []string{"text/html"}}).shouldReturnMediatype(t, "text/html")
}

func TestWildcardAcceptHeaderAndSupportedTypesHasOneElement_ReturnsSupportedType(t *testing.T) {
	(&negotiateWith{acceptHeader: "*/*", supportedTypes: []string{"text/html"}}).shouldReturnMediatype(t, "text/html")
	(&negotiateWith{acceptHeader: "text/*", supportedTypes: []string{"text/html"}}).shouldReturnMediatype(t, "text/html")
}

func TestOnlyHTMLRequestedAndSupportedTypesContainsHTML_ReturnsHTML(t *testing.T) {
	(&negotiateWith{acceptHeader: "text/html", supportedTypes: []string{"application/json", "text/html"}}).shouldReturnMediatype(t, "text/html")
	(&negotiateWith{acceptHeader: "text/html; q=0.2", supportedTypes: []string{"application/json", "text/html"}}).shouldReturnMediatype(t, "text/html")
}

func TestMimeTypeAndSubtypeRequestedAndSubtypeSupported_ReturnsSubtype(t *testing.T) {
	(&negotiateWith{acceptHeader: "audio/*; q=0.2, audio/basic", supportedTypes: []string{"audio/basic"}}).shouldReturnMediatype(t, "audio/basic")
	(&negotiateWith{acceptHeader: "audio/*; q=0.2, audio/basic", supportedTypes: []string{"audio/basic", "audio/mp3"}}).shouldReturnMediatype(t, "audio/basic")
	(&negotiateWith{acceptHeader: "audio/basic, audio/*; q=0.2", supportedTypes: []string{"audio/basic", "audio/mp3"}}).shouldReturnMediatype(t, "audio/basic")
}

func TestMimeTypeAndSubtypeRequestedAndOtherSubtypeSupported_ReturnsOtherSubtype(t *testing.T) {
	(&negotiateWith{acceptHeader: "audio/*; q=0.2, audio/basic", supportedTypes: []string{"audio/mp3"}}).shouldReturnMediatype(t, "audio/mp3")
}

func TestAcceptContainsUnknownValues_ValuesAreIgnored(t *testing.T) {
	(&negotiateWith{acceptHeader: "audio/*; q=0.2, audio/basic; level=1", supportedTypes: []string{"audio/basic", "audio/mp3"}}).shouldReturnMediatype(t, "audio/basic")
	(&negotiateWith{acceptHeader: "audio/*; q=0.2, audio/basic; q=bla", supportedTypes: []string{"audio/basic", "audio/mp3"}}).shouldReturnMediatype(t, "audio/basic")
}

func TestBothRequestedTypesAreSupported_ReturnsTypeWithHigherWeight(t *testing.T) {
	(&negotiateWith{acceptHeader: "text/html, application/json; q=0.2", supportedTypes: []string{"text/html", "application/json"}}).shouldReturnMediatype(t, "text/html")
	(&negotiateWith{acceptHeader: "text/html; q=0.1, application/json; q=0.2", supportedTypes: []string{"text/html", "application/json"}}).shouldReturnMediatype(t, "application/json")
	(&negotiateWith{acceptHeader: "text/html; q=0.2, application/json; q=0.1", supportedTypes: []string{"text/html", "application/json"}}).shouldReturnMediatype(t, "text/html")
}

func TestComplexCombinations(t *testing.T) {
	(&negotiateWith{acceptHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", supportedTypes: []string{"text/html", "application/json"}}).shouldReturnMediatype(t, "text/html")
	(&negotiateWith{acceptHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", supportedTypes: []string{"application/json"}}).shouldReturnMediatype(t, "application/json")
	(&negotiateWith{acceptHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", supportedTypes: []string{"application/json", "application/xml"}}).shouldReturnMediatype(t, "application/xml")
	(&negotiateWith{acceptHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8", supportedTypes: []string{"application/json", "application/xml", "image/webp"}}).shouldReturnMediatype(t, "image/webp")
}

type negotiateWith struct {
	acceptHeader   string
	supportedTypes []string
}

func (input *negotiateWith) shouldReturnErrorNotSupported(t *testing.T) {
	_, err := mediatype.Negotiate(input.acceptHeader, input.supportedTypes)
	if err != mediatype.ErrNotSupported {
		t.Error("Expected to get ErrNotSupported but got ", err)
	}
}

func (input *negotiateWith) shouldReturnMediatype(t *testing.T, expectedMediatype string) {
	m, err := mediatype.Negotiate(input.acceptHeader, input.supportedTypes)

	if err != nil {
		t.Fatalf("Negotiate(%v) with 'Accept:%v': expected %v but got %v", input.supportedTypes, input.acceptHeader, expectedMediatype, err)
	}

	if m.String() != expectedMediatype {
		t.Fatalf("Negotiate(%v) with 'Accept:%v': expected %v but got %v", input.supportedTypes, input.acceptHeader, expectedMediatype, m)
	}
}
