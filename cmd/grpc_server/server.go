package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/auth-reverse-proxy/grpcserver"
)

var (
	port = flag.Uint("port", 8080, "Port to listen to")
)

func main() {
	srv := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		panic(err)
	}
	grpcserver.RegisterPingServiceServer(srv, grpcserver.DefaultPingServiceServer)

	errs := make(chan error)

	go func() {
		log.Printf("listening on %s", lis.Addr().String())
		errs <- srv.Serve(lis)
	}()

	// Intersept signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)


	// Signal handler
	go func() {
		sig := <-sigs
		log.Printf("shutdown due to %s", sig)
		srv.GracefulStop()
	}()

	if err := <-errs; err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	flag.Parse()
}