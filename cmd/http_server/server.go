package main

import (
	"errors"
	"io"
	"net/http"
	"k8s.io/klog/v2"
	"os"
)

func introspectRequest(r *http.Request){
	klog.Info("Method: ", r.Method)
    klog.Info("URL: ", r.URL)
    klog.Info("Header: ", r.Header)
    klog.Info("Content-Type: ", r.Header.Get("Content-Type"))

	body, _ := io.ReadAll(r.Body)
    r.Body.Close()
    klog.Info("Body: ", body)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	klog.Info("got / request\n")
	introspectRequest(r)
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	klog.Info("got /hello request\n")
	introspectRequest(r)
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)

	err := http.ListenAndServe(":9091", nil)
	if errors.Is(err, http.ErrServerClosed) {
		klog.Info("server closed\n")
	} else if err != nil {
		klog.Info("error starting server: ", err)
		os.Exit(1)
	}
}