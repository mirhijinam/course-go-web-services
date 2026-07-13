package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
)

const (
	consumerKey = "consumer"
)

func gracefulStart(ctx context.Context, l net.Listener, srv *grpc.Server) {
	errCh := make(chan error)

	go func() {
		errCh <- srv.Serve(l)
	}()

	select {
	case <-ctx.Done():
		srv.GracefulStop()
		<-errCh
	case err := <-errCh:
		if err != nil {
			panic(fmt.Sprintf("failed to serve %v: %v", l.Addr(), err))
		}
	}
}

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) error {
	var acl map[string][]string
	err := json.Unmarshal([]byte(ACLData), &acl)
	if err != nil {
		return err
	}

	listen, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	clients := NewClientQueues()
	// Создаем интерцептор с acl
	authInterceptor := NewAuthInterceptor(acl)
	distributionInterceptor := NewDistributionInterceptor(&clients)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary,
			distributionInterceptor.Unary,
		),
		grpc.ChainStreamInterceptor(
			authInterceptor.Stream,
			distributionInterceptor.Stream,
		),
	)

	RegisterAdminServer(server, NewAdmin(&clients))
	RegisterBizServer(server, NewBiz())

	go gracefulStart(ctx, listen, server)

	return nil
}

func GetConsumer(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	consumers, ok := md[consumerKey]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if len(consumers) == 0 {
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}

	consumer := consumers[0]

	return consumer, nil
}
