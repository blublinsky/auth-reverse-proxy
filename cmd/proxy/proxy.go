package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"github.com/auth-reverse-proxy/auth"

)


func main() {
	// Remote server
	remote, err := url.Parse("http://localhost:9091")
	if err != nil {
			panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
//	http.HandleFunc("/", auth.TokenAuthfunc(proxy, remote))
//	http.HandleFunc("/", auth.BasicAuthfunc(proxy, remote))
	http.HandleFunc("/signon", auth.JWTSignOn())
	http.HandleFunc("/", auth.JWTAuthfunc(proxy, remote))
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		panic(err)
	}
}
