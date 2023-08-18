package auth

import (
	"crypto/subtle"
	"net/http"
	"net/http/httputil"
	"net/url"

	"k8s.io/klog/v2"
)

const (
	user    = "boris"
	passwrd = "12345"
)

// This is a very simple implementtion that is based on pre defined user name/password. In reality, this
// Requires integration with some user management sytem, whether it is an LDAP, etc...
func BasicAuthfunc(p *httputil.ReverseProxy, upstream *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Destination URL", r.URL)
		// Get user/password from request
		usr, pass, ok := r.BasicAuth()
		if !ok {
			klog.Info("Failed to obtain credentials")
			WriteUnauthorisedResponse(w)
			return
		}
		klog.Info("submitted credentials: user - ", usr, " password - ", pass)
		if subtle.ConstantTimeCompare([]byte(usr), []byte(user)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(passwrd)) != 1 {
			// Wrong user/password
			WriteUnauthorisedResponse(w)
			return
		}
		modifyRequest(r, upstream)
		p.ServeHTTP(w, r)
	}
}
