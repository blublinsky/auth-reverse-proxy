package auth

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type authorization struct {
	proxy    *httputil.ReverseProxy
	prefix   string
	upstream *url.URL
}

// Create Unauthorised response
func WriteUnauthorisedResponse(w http.ResponseWriter) {
	w.WriteHeader(401)
	w.Write([]byte("Unauthorised\n"))
}

// Create Unauthorised response
func WriteBadRequestResponse(w http.ResponseWriter) {
	w.WriteHeader(400)
	w.Write([]byte("Bad Request\n"))
}

// Create Unauthorised response
func WriteInternalErrorResponse(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("Internal Server Error\n"))
}

// Modify request upstream URL
func modifyRequest(r *http.Request, upstream *url.URL) {
	r.URL.Host = upstream.Host
	r.URL.Scheme = upstream.Scheme
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Host = upstream.Host
}
