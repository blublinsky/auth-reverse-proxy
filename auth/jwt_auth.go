package auth

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/auth-reverse-proxy/auth/jwt"
	"k8s.io/klog/v2"
)

const (
	jwt_header = "jwt-token"
)


func JWTAuthfunc(p *httputil.ReverseProxy, upstream *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Destination URL", r.URL)
		// Get token
		auth := r.Header.Get(jwt_header)
		klog.Info("Auth token ", auth)

		// Validate token
		err := jwt.ValidateTokenString(auth)
		if err == 2 {
			// Bad request
			WriteBadRequestResponse(w)
			return
		}
		if err == 1 {
			// Unauthorised
			WriteUnauthorisedResponse(w)
			return
		}
		modifyRequest(r, upstream)
		p.ServeHTTP(w, r)
	}
}

// User signon
func JWTSignOn() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Info("New SignOn")
		user := jwt.ValidateUser(r)
		if user == nil {
			// Bad signon request
			WriteBadRequestResponse(w)
			return
		}	
		// Build new JWT token
		tkn := jwt.NewJWTTokenString(*user)
		if tkn == nil {
			WriteInternalErrorResponse(w)
			return
		}
		// return token
		w.WriteHeader(200)
		w.Write([]byte(*tkn))
	}
}