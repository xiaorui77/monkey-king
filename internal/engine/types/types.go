package types

import (
	"net/url"
	"strings"
)

type ResponseWarp struct {
	StatusCode int
	Body       []byte
	Request    *RequestWarp
}

type RequestWarp struct {
	URL     *url.URL
	Method  string
	BaseURL *url.URL
}

// AbsoluteURL 根据相对路径获取完整url
func (r *RequestWarp) AbsoluteURL(u string) string {
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
