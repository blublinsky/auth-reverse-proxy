package auth

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"k8s.io/klog/v2"
)

const (
	token_header = "auth-token"
	auth_token   = "12345"
)

func TokenAuthfunc(p *httputil.ReverseProxy, upstream *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Destination URL ", r.URL)
		auth := r.Header.Get(token_header)
		klog.Info("Auth token ", auth)
		if auth != auth_token {
			// Wrong token
			WriteUnauthorisedResponse(w)
			return
		}
		modifyRequest(r, upstream)
		p.ServeHTTP(w, r)
	}
}
