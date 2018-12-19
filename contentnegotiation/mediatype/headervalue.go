package mediatype

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	ErrSyntax = errors.New("headervalue: has wrong syntax")

	withQualityRegexp, _ = regexp.Compile("^\\s*([^\\s;]+)(\\s*;\\s*q=(\\d?[.]\\d+))*")
)

type headervalue struct {
	Value   string
	Quality float64
}

func parseHeaderValue(s string) (*headervalue, error) {
	h := withQualityRegexp.FindStringSubmatch(s)

	if h == nil {
		return nil, ErrSyntax
	}

	if h[2] != "" {
		q, err := strconv.ParseFloat(h[3], 64)

		if err != nil {
			return nil, ErrSyntax
		}
		return &headervalue{Value: h[1], Quality: q}, nil
	}
	return &headervalue{Value: h[1], Quality: 1.0}, nil
}

type headervalues []*headervalue

func (slice headervalues) Len() int {
	return len(slice)
}

func (slice headervalues) Less(i, j int) bool {
	return slice[i].Quality < slice[j].Quality
}

func (slice headervalues) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
