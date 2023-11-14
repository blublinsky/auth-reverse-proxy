package grpcserver

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	returnHeader = "test-client-header"
)

// TestPingServiceServerImpl can be called to test the underlying TestServiceServer.
func TestPingServiceServerImpl(t *testing.T, client PingServiceClient) {
	t.Run("Unary ping", func(t *testing.T) {
		want := "hello, world"
		hdr := metadata.MD{}
		res, err := client.Ping(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Authorization", "12345")), &PingRequest{Value: want}, grpc.Header(&hdr))
		if err != nil {
			t.Errorf("want no err; got %v", err)
			return
		}
		checkHeaders(t, hdr)
		t.Logf("got %v (%d)", res.Value, res.Counter)
		if got := res.Value; got != want {
			t.Errorf("res.Value = %q; want %q", got, want)
		}
	})

	t.Run("Unauthenticated ping", func(t *testing.T) {
		_, err := client.PingError(context.TODO(), &PingRequest{})
		if err == nil {
			t.Errorf("want err; got %v", err)
		}
		if err.Error() != "rpc error: code = Unauthenticated desc = Request unauthorised" {
			t.Errorf("want unauthentication err; got %v", err.Error())
		}
	})

	t.Run("Wrong input ping", func(t *testing.T) {
		_, err := client.PingError(metadata.NewOutgoingContext(context.Background(), metadata.Pairs("Authorization", "12345")), &PingRequest{})
		if err == nil {
			t.Errorf("want err; got %v", err)
		}
		if err.Error() != "rpc error: code = Unknown desc = Something is wrong and this is a message that describes it" {
			t.Errorf("want wrong input err; got %v", err.Error())
		}
	})

	t.Run("Unary ping with headers", func(t *testing.T) {
		want := "hello, world"
		req := &PingRequest{Value: want}

		ctx := metadata.AppendToOutgoingContext(context.Background(), returnHeader, "I like turtles.", "Authorization", "12345")
		inHeader := make(metadata.MD)

		res, err := client.Ping(ctx, req, grpc.Header(&inHeader))
		if err != nil {
			t.Errorf("want no err; got %v", err)
			return
		}
		t.Logf("got %v (%d)", res.Value, res.Counter)
		if !reflect.DeepEqual(inHeader.Get(returnHeader), []string{"I like turtles."}) {
			t.Errorf("did not receive correct return headers")
		}
	})
}

func checkHeaders(t *testing.T, md metadata.MD) {
	vs := md.Get(PingHeader)
	if want, got := 1, len(vs); want != got {
		t.Errorf("header %q not present", PingHeader)
		return
	}
	if want, got := []string{PingHeaderCts}, vs; !reflect.DeepEqual(got, want) {
		t.Errorf("header mismatch; want %q, got %q", want, got)
	}
}