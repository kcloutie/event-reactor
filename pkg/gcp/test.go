package gcp

import (
	"context"
	"fmt"
	"net"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FakeSecretManagerServerResponse struct {
	Response *secretmanagerpb.AccessSecretVersionResponse
	Err      error
}

type FakeSecretManagerServer struct {
	secretmanagerpb.UnimplementedSecretManagerServiceServer
	Responses map[string]FakeSecretManagerServerResponse
}

func (s *FakeSecretManagerServer) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {

	resp, exists := s.Responses[req.Name]
	if !exists {
		return nil, fmt.Errorf("secret '%s' not found", req.Name)
	}
	return resp.Response, resp.Err
	// req.Name = "projects/test/secrets/test/versions/latest"
	//
	//	return &secretmanagerpb.AccessSecretVersionResponse{
	//		Name:    req.Name,
	//		Payload: &secretmanagerpb.SecretPayload{Data: []byte("fake-secret")},
	//	}, nil
}

func NewFakeServerAndClient(ctx context.Context, t *testing.T) (*FakeSecretManagerServer, *secretmanager.Client) {
	server := &FakeSecretManagerServer{
		Responses: map[string]FakeSecretManagerServerResponse{},
	}

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer()
	secretmanagerpb.RegisterSecretManagerServiceServer(gsrv, server)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()

	client, err := secretmanager.NewClient(ctx,
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		// WithTransportCredentials and insecure.NewCredentials()
	)
	if err != nil {
		t.Fatal(err)
	}

	return server, client
}
