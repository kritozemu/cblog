package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(firstClient, secondClient))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, "label", "vip")
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 456,
	})
	assert.NoError(t, err)
	t.Log(resp)
}

var firstClient grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("这是第一个client前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("这是第一个client后")
	return err
}

var secondClient grpc.UnaryClientInterceptor = func(ctx context.Context,
	method string, req, reply any, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("这是第一个client前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("这是第二个client后")
	return err
}
