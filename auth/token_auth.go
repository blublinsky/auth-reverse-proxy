package auth

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"io"

	"k8s.io/klog/v2"
)

const (
	token_header = "authorization"
)

type tokenAuth struct {
	authorization
	token string
}

func introspectRequest(r *http.Request){
	klog.Info("Method: ", r.Method)
    klog.Info("URL: ", r.URL)
    klog.Info("Header: ", r.Header)
    klog.Info("Content-Type: ", r.Header.Get("Content-Type"))

	body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()
    klog.Info("Body: ", body)
}

func NewTokenAuth(token string, proxy *httputil.ReverseProxy, prefix string, upstream *url.URL) tokenAuth {
	auth := authorization{proxy: proxy, prefix: prefix, upstream: upstream}
	return tokenAuth{token: token, authorization: auth}
}

func (ta tokenAuth) AuthFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Destination URL ", r.URL)
		if strings.HasPrefix(r.URL.String(), ta.prefix) {
			auth := r.Header.Get(token_header)
			klog.Info("Auth token ", auth)
			if auth != ta.token {
				// Wrong token
				WriteUnauthorisedResponse(w)
				return
			}
		}
		modifyRequest(r, ta.upstream)
		ta.proxy.ServeHTTP(w, r)
	}
}
