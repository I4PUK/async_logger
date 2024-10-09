package main

import "context"

func getBizInstance() *BizServ {
	return &BizServ{}
}

type BizServ struct{}

func (biz *BizServ) Check(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz *BizServ) Add(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz *BizServ) Test(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz *BizServ) mustEmbedUnimplementedBizServer() {}
