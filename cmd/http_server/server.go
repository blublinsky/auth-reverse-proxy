package main

import (
	"errors"
	"io"
	"net/http"
	"k8s.io/klog/v2"
	"os"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	klog.Info("got / request\n")
	io.WriteString(w, "This is my website!\n")
}
func getHello(w http.ResponseWriter, r *http.Request) {
	klog.Info("got /hello request\n")
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