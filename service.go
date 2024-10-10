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
	"sync"
)

var (
	authKey            = "consumer"
	defaultHost        = "127.0.0.1"
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidConsumer = status.Errorf(codes.Unauthenticated, "invalid consumer")
)

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

func authorize(ctx context.Context, method string, acl map[string][]string) (bool, error) {
	consumer, err := getConsumerName(ctx)
	if err != nil {
		return false, err
	}

	methods, ok := acl[consumer]
	if !ok || len(methods) == 0 {
		return false, errInvalidConsumer
	}
	for _, m := range methods {
		fullM := strings.Split(m, "/")
		if strings.HasSuffix(method, fullM[1]+"/"+fullM[2]) || (fullM[2] == "*" && strings.Contains(method, fullM[1])) {
			return true, nil
		}
	}
	return false, nil
}

func streamAuthInterceptor(acl map[string][]string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ok, err := authorize(ss.Context(), info.FullMethod, acl)
		if err != nil {
			return err
		}
		if !ok {
			return errInvalidConsumer
		}
		return handler(srv, ss)
	}
}

func authInterceptor(acl map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ok, err := authorize(ctx, info.FullMethod, acl)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errInvalidConsumer
		}

		return handler(ctx, req)
	}
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

	logger := &SimpleEventLogger{
		mu:          sync.Mutex{},
		subscribers: make(map[chan *Event]struct{}),
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor(acl)), grpc.StreamInterceptor(streamAuthInterceptor(acl)))

	bizModule := getBizInstance()
	adminModule := getAdminInstance(logger)

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
