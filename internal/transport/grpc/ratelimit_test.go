package grpc

import (
	"context"
	"testing"

	"go-usersvc-demo/internal/pkg/ratelimit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestUnaryRateLimitInterceptor_Allow(t *testing.T) {
	limiter := ratelimit.NewLimiter(100, 10)
	defer limiter.ResetAll()

	interceptor := UnaryRateLimitInterceptor(limiter)

	// Mock peer info
	p := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "127.0.0.1:50051",
		},
	}
	ctx := peer.NewContext(context.Background(), p)

	var handlerCalled bool
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		Server:     nil,
		FullMethod: "/pb.UserService/CreateUser",
	}

	_, err := interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestUnaryRateLimitInterceptor_RateLimited(t *testing.T) {
	// Very strict limit
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	interceptor := UnaryRateLimitInterceptor(limiter)

	p := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "127.0.0.1:50051",
		},
	}
	ctx := peer.NewContext(context.Background(), p)

	info := &grpc.UnaryServerInfo{
		Server:     nil,
		FullMethod: "/pb.UserService/CreateUser",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// First call - should succeed
	_, err1 := interceptor(ctx, nil, info, handler)
	if err1 != nil {
		t.Errorf("Expected first call to succeed, got %v", err1)
	}

	// Second call - should be rate limited
	_, err2 := interceptor(ctx, nil, info, handler)
	if err2 == nil {
		t.Error("Expected second call to be rate limited")
	}

	// Check error code
	st, ok := status.FromError(err2)
	if !ok {
		t.Fatalf("Expected grpc status error, got %v", err2)
	}
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("Expected ResourceExhausted code, got %v", st.Code())
	}
}

func TestUnaryRateLimitInterceptor_DifferentIPs(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	interceptor := UnaryRateLimitInterceptor(limiter)

	info := &grpc.UnaryServerInfo{
		Server:     nil,
		FullMethod: "/pb.UserService/CreateUser",
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// Request from IP1
	p1 := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "192.168.1.1:50051",
		},
	}
	ctx1 := peer.NewContext(context.Background(), p1)

	_, err1 := interceptor(ctx1, nil, info, handler)
	if err1 != nil {
		t.Errorf("Expected IP1 request to succeed, got %v", err1)
	}

	// Request from IP2
	p2 := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "192.168.1.2:50051",
		},
	}
	ctx2 := peer.NewContext(context.Background(), p2)

	_, err2 := interceptor(ctx2, nil, info, handler)
	if err2 != nil {
		t.Errorf("Expected IP2 request to succeed (different IP), got %v", err2)
	}
}

func TestStreamRateLimitInterceptor_Allow(t *testing.T) {
	limiter := ratelimit.NewLimiter(100, 10)
	defer limiter.ResetAll()

	interceptor := StreamRateLimitInterceptor(limiter)

	p := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "127.0.0.1:50051",
		},
	}
	ctx := peer.NewContext(context.Background(), p)

	ss := &MockServerStream{
		ctx: ctx,
	}

	info := &grpc.StreamServerInfo{

		FullMethod:     "/pb.UserService/ListUsers",
		IsClientStream: false,
		IsServerStream: true,
	}

	var handlerCalled bool
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		handlerCalled = true
		return nil
	}

	err := interceptor(nil, ss, info, handler)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestStreamRateLimitInterceptor_RateLimited(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	interceptor := StreamRateLimitInterceptor(limiter)

	p := &peer.Peer{
		Addr: &MockAddr{
			network: "tcp",
			addr:    "127.0.0.1:50051",
		},
	}
	ctx := peer.NewContext(context.Background(), p)

	info := &grpc.StreamServerInfo{
		FullMethod: "/pb.UserService/ListUsers",
	}

	handler := func(srv interface{}, ss grpc.ServerStream) error {
		return nil
	}

	// First call
	ss1 := &MockServerStream{ctx: ctx}
	err1 := interceptor(nil, ss1, info, handler)
	if err1 != nil {
		t.Errorf("Expected first call to succeed, got %v", err1)
	}

	// Second call - rate limited
	ss2 := &MockServerStream{ctx: ctx}
	err2 := interceptor(nil, ss2, info, handler)
	if err2 == nil {
		t.Error("Expected second call to be rate limited")
	}

	st, ok := status.FromError(err2)
	if !ok {
		t.Fatalf("Expected grpc status error, got %v", err2)
	}
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("Expected ResourceExhausted code, got %v", st.Code())
	}
}

// Mock implementations for testing
type MockAddr struct {
	network string
	addr    string
}

func (m *MockAddr) Network() string {
	return m.network
}

func (m *MockAddr) String() string {
	return m.addr
}

type MockServerStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (m *MockServerStream) Context() context.Context {
	return m.ctx
}
