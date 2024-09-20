package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type BizModule struct{}

func (biz BizModule) mustEmbedUnimplementedBizServer() {
	//TODO implement me
	panic("implement me")
}

func getBizInstance() *BizModule {
	return &BizModule{}
}

func (biz BizModule) Check(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz BizModule) Add(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz BizModule) Test(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

type AdminModule struct {
	statisticsData ServiceStat
	logData        ServiceLogger
}

func (a AdminModule) mustEmbedUnimplementedAdminServer() {
	//TODO implement me
	panic("implement me")
}

func (a AdminModule) Logging(n *Nothing, logServer Admin_LoggingServer) error {
	return nil
}

func (a AdminModule) Statistics(interval *StatInterval, statServer Admin_StatisticsServer) error {
	return nil
}

func getAdminInstance() *AdminModule {
	return &AdminModule{}
}

type ServiceLogger struct {
}

type ServiceStat struct {
}

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
func StartMyMicroservice(ctx context.Context, addr string, data string) error {
	var acl map[string]interface{}

	err := json.Unmarshal([]byte(data), &acl)
	if err != nil {
		log.Println(err)
		return err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Cannot listen port: ", err)
	}

	server := grpc.NewServer()

	bizModule := getBizInstance()
	adminModule := getAdminInstance()

	RegisterBizServer(server, bizModule)
	RegisterAdminServer(server, adminModule)

	fmt.Println("starting server at ", addr)

	go server.Serve(listener)
	go ServerStopper(ctx, server)

	return nil
}

func ServerStopper(ctx context.Context, server *grpc.Server) {
	for {
		select {
		case <-ctx.Done():
			server.Stop()
			return
		}
	}
}
