package main

import (
	"context"
	"net"
	"k8s.io/klog/v2"


	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/auth-reverse-proxy/grpcproxy"
)

func main() {
	//  Testing values
	remote_port := "8080"
	local_port := "8081"
	security_token := "12345"

	// Client connection
	cc, err := grpc.Dial("localhost:" + remote_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatal("cannot dial server: ", err)
	}
	defer cc.Close()
	klog.Info("connecting to server ", cc.Target())

	// set up listener
	lis, err := net.Listen("tcp", "localhost:" + local_port)
	if err != nil {
		klog.Fatalf("failed to listen: ", err)
	}
	klog.Info("listening on ", lis.Addr())

	// Director function
	directorFn := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
		return outCtx, cc, nil
	}

	// Set up the proxy server and then serve from it like in step one.
	sh, h := grpcproxy.TransparentHandler(directorFn)
	proxySrv := grpc.NewServer(grpc.UnknownServiceHandler(sh))
	// Add security header checking
	h.AddSecurityHeaderToHandler(map[string]string{ "Authorization": security_token})
	klog.Info("Security header added ")

	// run the proxy backend
	klog.Info("Running proxySrv")
	if err := proxySrv.Serve(lis); err != nil {
		klog.Fatal("server stopped unexpectedly")
	} 
}