package main

import (
	"context"
	"k8s.io/klog/v2"


	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/auth-reverse-proxy/grpcserver"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	//  Testing values
	proxy_port := "8081"
	security_token := "12345"

	// Client connection
	cc, err := grpc.Dial("localhost:" + proxy_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatal("cannot dial server: ", err)
	}
	defer cc.Close()
	klog.Info("connecting to server ", cc.Target())

	// create client
	client := grpcserver.NewPingServiceClient(cc)

	// Send some requests
	_, err = client.PingEmpty(context.Background(), &emptypb.Empty{})
	klog.Info("unauthenticated request, expected error, got ", err.Error())

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Authorization", security_token))
	out, err := client.PingEmpty(ctx, &emptypb.Empty{})
	if err != nil {
		klog.Error("Expected error, got ", err.Error())
	} else {
		klog.Info("get result ", out)
	}

	out, err = client.Ping(ctx, &grpcserver.PingRequest{Value: "foo"})
	if err != nil {
		klog.Error("Expected error, got ", err.Error())
	} else {
		klog.Info("get result ", out)
	}
}