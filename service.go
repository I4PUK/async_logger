package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
	"time"
)

var (
	authKey            = "consumer"
	defaultHost        = "127.0.0.1"
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidConsumer = status.Errorf(codes.Unauthenticated, "invalid consumer")
)

var ServiceEventLogger SimpleEventLogger

type BizServ struct{}
type AdminServ struct{}

func (biz BizServ) mustEmbedUnimplementedBizServer() {}

func getBizInstance() *BizServ {
	return &BizServ{}
}

func (biz BizServ) Check(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz BizServ) Add(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (biz BizServ) Test(ctx context.Context, nthg *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (adm *AdminServ) mustEmbedUnimplementedAdminServer() {}

func (adm *AdminServ) Logging(n *Nothing, logServerStream Admin_LoggingServer) error {
	ch := ServiceEventLogger.Subscribe()

	for e := range ch {
		err := logServerStream.Send(e)
		if err != nil {
			ServiceEventLogger.Unsubscribe(ch)
			return err
		}
	}
	return nil
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

func emptyStat() *Stat {
	return &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
}

func updateStat(stat *Stat, e *Event) {
	stat.Timestamp = time.Now().Unix()
	stat.ByConsumer[e.Consumer]++
	stat.ByMethod[e.Method]++
}

func (adm *AdminServ) Statistics(interval *StatInterval, stream Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	ch := ServiceEventLogger.Subscribe()
	stat := emptyStat()

	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			ServiceEventLogger.Unsubscribe(ch)
			return stream.Context().Err()
		case e := <-ch:
			updateStat(stat, e)
		case <-ticker.C:
			err := stream.Send(stat)
			if err != nil {
				ServiceEventLogger.Unsubscribe(ch)
				return err
			}
			stat = emptyStat()
		}
	}
}

func getAdminInstance() *AdminServ {
	return &AdminServ{}
}

// StartMyMicroservice начальная точка входа
func StartMyMicroservice(ctx context.Context, addr string, ACLData string) error {
	acl := make(map[string][]string)

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

	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Println("Cannot accept connection: ", err)
		}
	}()
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
