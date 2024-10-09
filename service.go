package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	authKey            = "consumer"
	defaultHost        = "127.0.0.1"
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidConsumer = status.Errorf(codes.Unauthenticated, "invalid consumer")
)

type EventLogger interface {
	LogEvent(consumer, method, host string)
	Subscribe() chan *pb.Event
	Unsubscribe(chan *pb.Event)
}

type SimpleEventLogger struct {
	mu          sync.Mutex
	subscribers map[chan *pb.Event]struct{}
}

type BizModule struct{}

func (biz BizModule) mustEmbedUnimplementedBizServer() {}

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

func timingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	fmt.Printf("--"+
		"call=%v"+
		"req=%#v"+
		"reply=%#v"+
		"time=%#v"+
		"err=%#v", time.Since(start))
	return nil, nil
}

func getConsumerName(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}
	consumers, ok := md[authKey]
	if !ok || len(consumers) == 0 {
		return "", errInvalidConsumer
	}
	return consumers[0], nil
}

func authInterceptor(acl map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		consumer, err := getConsumerName(ctx)
		method := info.FullMethod

		if err != nil {
			return nil, err
		}

		methods, ok := acl[consumer]
		if !ok || len(methods) == 0 {
			return nil, errInvalidConsumer
		}
		for _, m := range methods {
			fullM := strings.Split(m, "/")
			if strings.HasSuffix(method, fullM[1]+"/"+fullM[2]) || (fullM[2] == "*" && strings.Contains(method, fullM[1])) {
				return handler(ctx, req)
			}
		}

		return nil, errInvalidConsumer
	}
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

	err := json.Unmarshal([]byte(ACLData), &acl)
	if err != nil {
		log.Println("Invalid ACL data: " + ACLData)
		return err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Cannot listen port: ", err)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor(acl)))

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
