package main

import (
	"context"
)

type Biz struct {
	UnimplementedBizServer
}

func NewBiz() Biz {
	return Biz{
		UnimplementedBizServer: UnimplementedBizServer{},
	}
}

func (b Biz) Check(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b Biz) Add(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b Biz) Test(ctx context.Context, in *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}
