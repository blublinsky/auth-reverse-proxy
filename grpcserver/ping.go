package grpcserver

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var DefaultPingServiceServer = defaultPingServer{}

const (
	PingHeader      = "ping-header"
	PingHeaderCts   = "Arbitrary header text"
	PingTrailer     = "ping-trailer"
	PingTrailerCts  = "Arbitrary trailer text"
	PingEchoHeader  = "ping-echo-header"
	PingEchoTrailer = "ping-echo-trailer"
)

// defaultPingServer is the canonical implementation of a Ping server.
type defaultPingServer struct {
	UnimplementedPingServiceServer
}

func (s defaultPingServer) PingEmpty(ctx context.Context, empty *emptypb.Empty) (*PingResponse, error) {
	if err := s.sendHeader(ctx); err != nil {
		return nil, err
	}
	if err := s.setTrailer(ctx); err != nil {
		return nil, err
	}
	return &PingResponse{}, nil
}

func (s defaultPingServer) Ping(ctx context.Context, request *PingRequest) (*PingResponse, error) {
	if err := s.sendHeader(ctx); err != nil {
		return nil, err
	}
	if err := s.setTrailer(ctx); err != nil {
		return nil, err
	}

	return &PingResponse{Value: request.Value}, nil
}

func (s defaultPingServer) PingError(ctx context.Context, request *PingRequest) (*emptypb.Empty, error) {
	if err := s.sendHeader(ctx); err != nil {
		return nil, err
	}
	if err := s.setTrailer(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unknown, "Something is wrong and this is a message that describes it")
}


// Support methods
func (s *defaultPingServer) sendHeader(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	if tvs := md.Get(PingEchoHeader); len(tvs) > 0 {
		md.Append(PingEchoHeader, tvs...)
	}

	md.Append(PingHeader, PingHeaderCts)

	if err := grpc.SendHeader(ctx, md); err != nil {
		return fmt.Errorf("setting header: %w", err)
	}
	return nil
}

func (s *defaultPingServer) setTrailer(ctx context.Context) error {
	md := s.buildTrailer(ctx)

	if err := grpc.SetTrailer(ctx, md); err != nil {
		return fmt.Errorf("setting trailer: %w", err)
	}

	return nil
}

func (s *defaultPingServer) buildTrailer(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	if tvs := md.Get(PingEchoTrailer); len(tvs) > 0 {
		md.Append(PingEchoTrailer, tvs...)
	}

	md.Append(PingTrailer, PingTrailerCts)

	return md
}

var _ PingServiceServer = (*defaultPingServer)(nil)