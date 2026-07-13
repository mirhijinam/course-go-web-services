package main

import (
	"context"
	"strings"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	acl map[string][]string
}

func NewAuthInterceptor(acl map[string][]string) AuthInterceptor {
	return AuthInterceptor{acl: acl}
}

func (ai *AuthInterceptor) auth(ctx context.Context, method string) error {
	consumer, err := GetConsumer(ctx)
	if err != nil {
		return err
	}

	methodParts := strings.Split(method, "/")

	allowedMethods := ai.acl[consumer]
	if len(allowedMethods) == 0 {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	var methodAllowed bool
	for _, m := range allowedMethods {
		if m == method {
			methodAllowed = true
			break
		}

		mParts := strings.Split(m, "/")
		if len(mParts) != len(methodParts) {
			continue
		}

		isMatch := true
		for i, part := range mParts {
			if part != methodParts[i] && part != "*" {
				isMatch = false
				break
			}
		}

		if isMatch {
			methodAllowed = true
			break
		}
	}

	if !methodAllowed {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return nil
}

func (ai *AuthInterceptor) Unary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	err := ai.auth(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	reply, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (ai *AuthInterceptor) Stream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := ai.auth(ss.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	err = handler(srv, ss)
	if err != nil {
		return err
	}

	return nil
}
