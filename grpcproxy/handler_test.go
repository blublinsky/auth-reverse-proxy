package grpcproxy_test

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/auth-reverse-proxy/grpcproxy"
	pb "github.com/auth-reverse-proxy/grpcserver"
)

const (
	pingDefaultValue   = "I like kittens."
	clientMdKey        = "test-client-header"
	serverHeaderMdKey  = "test-client-header"
	serverTrailerMdKey = "test-client-trailer"

	rejectingMdKey = "test-reject-rpc-if-in-context"
)

// asserting service is implemented on the server side and serves as a handler for stuff
type assertingService struct {
	t *testing.T
	pb.UnimplementedPingServiceServer
}

var _ pb.PingServiceServer = (*assertingService)(nil)

func (s *assertingService) PingEmpty(ctx context.Context, _ *emptypb.Empty) (*pb.PingResponse, error) {
	// Check that this call has client's metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	assert.True(s.t, ok, "PingEmpty call must have metadata in context")
	_, ok = md[clientMdKey]
	assert.True(s.t, ok, "PingEmpty call must have clients's custom headers in metadata")
	return &pb.PingResponse{Value: pingDefaultValue, Counter: 42}, nil
}

func (s *assertingService) Ping(ctx context.Context, ping *pb.PingRequest) (*pb.PingResponse, error) {
	// Send user trailers and headers.
	grpc.SendHeader(ctx, metadata.Pairs(serverHeaderMdKey, "I like turtles."))
	grpc.SetTrailer(ctx, metadata.Pairs(serverTrailerMdKey, "I like ending turtles."))
	return &pb.PingResponse{Value: ping.Value, Counter: 42}, nil
}

func (s *assertingService) PingError(ctx context.Context, ping *pb.PingRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.FailedPrecondition, "Userspace error.")
}


// ProxyHappySuite tests the "happy" path of handling: that everything works in absence of connection issues.
type ProxyHappySuite struct {
	suite.Suite

	serverListener   net.Listener
	server           *grpc.Server
	proxyListener    net.Listener
	proxy            *grpc.Server
	serverClientConn *grpc.ClientConn

	client     *grpc.ClientConn
	testClient pb.PingServiceClient
}

func (s *ProxyHappySuite) TestPingEmptyCarriesClientMetadataNoAuthorization() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(clientMdKey, "true"))
	_, err := s.testClient.PingEmpty(ctx, &emptypb.Empty{})
	require.Error(s.T(), err, "PingEmpty should with error")
	e := err.Error()
	require.True(s.T(), e == "rpc error: code = Unauthenticated desc = Request unauthorised")
}

func (s *ProxyHappySuite) TestPingEmptyCarriesClientMetadata() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(clientMdKey, "true", "Authorization", "12345"))
	out, err := s.testClient.PingEmpty(ctx, &emptypb.Empty{})
	require.NoError(s.T(), err, "PingEmpty should succeed without errors")
	want := &pb.PingResponse{Value: pingDefaultValue, Counter: 42}
	require.True(s.T(), proto.Equal(want, out))
}

func (s *ProxyHappySuite) TestPingEmpty_StressTest() {
	for i := 0; i < 50; i++ {
		s.TestPingEmptyCarriesClientMetadata()
	}
}

func (s *ProxyHappySuite) TestPingCarriesServerHeadersAndTrailers() {
	headerMd := make(metadata.MD)
	trailerMd := make(metadata.MD)
	// This is an awkward calling convention... but meh.
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Authorization", "12345"))
	out, err := s.testClient.Ping(ctx, &pb.PingRequest{Value: "foo"}, grpc.Header(&headerMd), grpc.Trailer(&trailerMd))
	want := &pb.PingResponse{Value: "foo", Counter: 42}
	require.NoError(s.T(), err, "Ping should succeed without errors")
	require.True(s.T(), proto.Equal(want, out))
	assert.Contains(s.T(), headerMd, serverHeaderMdKey, "server response headers must contain server data")
	assert.Len(s.T(), trailerMd, 1, "server response trailers must contain server data")
}

func (s *ProxyHappySuite) TestPingErrorPropagatesAppError() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Authorization", "12345"))
	_, err := s.testClient.PingError(ctx, &pb.PingRequest{Value: "foo"})
	require.Error(s.T(), err, "PingError should never succeed")
	assert.Equal(s.T(), codes.FailedPrecondition, status.Code(err))
	assert.Equal(s.T(), "Userspace error.", status.Convert(err).Message())
}

func (s *ProxyHappySuite) TestDirectorErrorIsPropagated() {
	// See SetupSuite where the StreamDirector has a special case.
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(rejectingMdKey, "true", "Authorization", "12345"))
	_, err := s.testClient.Ping(ctx, &pb.PingRequest{Value: "foo"})
	require.Error(s.T(), err, "Director should reject this RPC")
	assert.Equal(s.T(), codes.PermissionDenied, status.Code(err))
	assert.Equal(s.T(), "testing rejection", status.Convert(err).Message())
}

func (s *ProxyHappySuite) SetupSuite() {
	var err error

	s.proxyListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for proxyListener")
	s.serverListener, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(s.T(), err, "must be able to allocate a port for serverListener")

	s.server = grpc.NewServer()
	pb.RegisterPingServiceServer(s.server, &assertingService{t: s.T()})

	// Setup of the proxy's Director.
	s.serverClientConn, err = grpc.Dial(s.serverListener.Addr().String(), grpc.WithInsecure())
	require.NoError(s.T(), err, "must not error on deferred client Dial")
	director := func(ctx context.Context, fullName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if _, exists := md[rejectingMdKey]; exists {
				return ctx, nil, status.Errorf(codes.PermissionDenied, "testing rejection")
			}
		}
		// Explicitly copy the metadata, otherwise the tests will fail.
		outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
		return outCtx, s.serverClientConn, nil
	}
	sh, h := grpcproxy.TransparentHandler(director)
	s.proxy = grpc.NewServer(grpc.UnknownServiceHandler(sh))
	h.AddSecurityHeaderToHandler(map[string]string{ "Authorization": "12345"})
	// Ping handler is handled as an explicit registration and not as a TransparentHandler.
	h = grpcproxy.RegisterService(s.proxy, director,
		"grpcserver.pingproto.PingService",
		"Ping")
	h.AddSecurityHeaderToHandler(map[string]string{ "Authorization": "12345"})

	// Start the serving loops.
	s.T().Logf("starting grpc.Server at: %v", s.serverListener.Addr().String())
	go func() {
		s.server.Serve(s.serverListener)
	}()
	s.T().Logf("starting grpc.Proxy at: %v", s.proxyListener.Addr().String())
	go func() {
		s.proxy.Serve(s.proxyListener)
	}()

	dCtx, ccl := context.WithTimeout(context.Background(), time.Second)
	defer ccl()
	ctx := metadata.NewOutgoingContext(dCtx, metadata.Pairs(rejectingMdKey, "true", "Authorization", "12345"))
	clientConn, err := grpc.DialContext(ctx, strings.Replace(s.proxyListener.Addr().String(), "127.0.0.1", "localhost", 1), grpc.WithInsecure())
	require.NoError(s.T(), err, "must not error on deferred client Dial")
	s.testClient = pb.NewPingServiceClient(clientConn)
}

func (s *ProxyHappySuite) TearDownSuite() {
	if s.client != nil {
		s.client.Close()
	}
	if s.serverClientConn != nil {
		s.serverClientConn.Close()
	}
	// Close all transports so the logs don't get spammy.
	time.Sleep(10 * time.Millisecond)
	if s.proxy != nil {
		s.proxy.Stop()
		s.proxyListener.Close()
	}
	if s.serverListener != nil {
		s.server.Stop()
		s.serverListener.Close()
	}
}

func TestProxyHappySuite(t *testing.T) {
	suite.Run(t, &ProxyHappySuite{})
}