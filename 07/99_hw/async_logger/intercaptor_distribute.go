package main

import (
	"context"
	"time"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	status "google.golang.org/grpc/status"
)

type DistributionInterceptor struct {
	queues *ClientQueues
}

func NewDistributionInterceptor(queues *ClientQueues) DistributionInterceptor {
	return DistributionInterceptor{
		queues: queues,
	}
}

func (di *DistributionInterceptor) distribute(e *Event) {
	di.queues.mu.Lock()
	defer di.queues.mu.Unlock()

	e.Timestamp = time.Now().Unix()
	for _, c := range di.queues.channels {
		c <- e
	}
}

func (di *DistributionInterceptor) Unary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	consumer, err := GetConsumer(ctx)
	if err != nil {
		return nil, err
	}

	addr, err := GetAddr(ctx)
	if err != nil {
		return nil, err
	}

	e := &Event{
		Consumer: consumer,
		Method:   info.FullMethod,
		Host:     addr,
	}

	di.distribute(e)

	reply, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (di *DistributionInterceptor) Stream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	consumer, err := GetConsumer(ctx)
	if err != nil {
		return err
	}

	addr, err := GetAddr(ctx)
	if err != nil {
		return err
	}

	e := &Event{
		Consumer: consumer,
		Method:   info.FullMethod,
		Host:     addr,
	}

	di.distribute(e)

	err = handler(srv, ss)
	if err != nil {
		return err
	}

	return nil
}

func GetAddr(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}
	return pr.Addr.String(), nil
}
