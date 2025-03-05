package demoserver

import (
	"context"
	"testing"

	pb "dev.azure.com/service-hub-flg/service_hub_validation/_git/service_hub_validation_service.git/mygreeterv3/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestSayHello(t *testing.T) {
	// Create a buffer connection
	bufSize := 1024 * 1024
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMyGreeterServer(s, &Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer s.Stop()

	// Create a client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewMyGreeterClient(conn)

	// Test SayHello method
	req := &pb.HelloRequest{Name: "Test", Age: 25, Email: "test@example.com"}
	resp, err := client.SayHello(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "Echo back what you sent me (SayHello): Test 25 test@example.com", resp.Message)
}
