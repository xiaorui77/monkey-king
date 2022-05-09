package types

import (
	"net/url"
	"strings"
)

type ResponseWarp struct {
	StatusCode int
	Body       []byte
	Request    *RequestWrap
}

type RequestWrap struct {
	URL     *url.URL
	Method  string
	BaseURL *url.URL
}

// AbsoluteURL return the absolute url according to the relative path.
func (r *RequestWrap) AbsoluteURL(u string) string {
	if strings.HasPrefix(u, "#") {
		return ""
	}
	var base *url.URL
	if r.BaseURL != nil {
		base = r.BaseURL
	} else {
		base = r.URL
	}
	absURL, err := base.Parse(u)
	if err != nil {
		return ""
	}
	absURL.Fragment = ""
	if absURL.Scheme == "//" {
		absURL.Scheme = r.URL.Scheme
	}
	return absURL.String()
}
