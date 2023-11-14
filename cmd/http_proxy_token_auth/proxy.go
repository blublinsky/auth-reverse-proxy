package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
//	"os"
	"github.com/auth-reverse-proxy/auth"
	"k8s.io/klog/v2"
)


func main() {
	// Get parameters
/*	remote_port := os.Getenv("REMOTE_PORT")
	local_port := os.Getenv("LOCAL_PORT")
	secure_prefix := os.Getenv("SECURE_PREFIX")
	security_token := os.Getenv("SECURITY_TOKEN")
*/
//  Testing values
	remote_port := "9091"
	local_port := "9090"
	secure_prefix := "/hello"
	security_token := "12345"


	if remote_port == "" || local_port == "" || secure_prefix == "" || security_token == "" {
		klog.Fatal("Failing to get execution parameters - remote port: ", remote_port, " local port: ", local_port, " secure prfix: ", secure_prefix)
	}
	klog.Info("Starting reverse proxy with parameters - remote port: ", remote_port, " local port: ", local_port, " secure prfix: ", secure_prefix)

	// Remote server
	remote, err := url.Parse("http://localhost:" + remote_port)
	if err != nil {
		klog.Fatal("Failed to parse remote url - error: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	token := auth.NewTokenAuth(security_token, proxy, secure_prefix, remote)

	http.HandleFunc("/", token.AuthFunc())
	err = http.ListenAndServe(":" + local_port, nil)
	if err != nil {
		klog.Fatal("HTTP server died unexpectidly, error - ", err)
	}
}
