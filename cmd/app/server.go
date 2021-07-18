package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/coomp/ccs/pkg/services"
	"github.com/coomp/ccs/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type MessageService struct {
}

func (svc *MessageService) OnEcho(ctx context.Context, in *services.EchoRequest) (*services.EchoResponse, error) {
	log.Printf("receive echo request: %s", in.Str)

	return &services.EchoResponse{
		Str: in.Str,
	}, nil
}

func (svc *MessageService) OnMessageRequest(ctx context.Context, in *services.MessageRequest) (*services.MessageResponse, error) {
	log.Printf("receive message request: [AppId => %s, ServiceId => %s]", in.AppId, in.ServiceId)

	// Token check

	// TODO get secretKey by AppId
	secretKey := "__SECRET_KEY__"

	var referers []*services.Referer

	token := util.HmacSha256Base64(fmt.Sprintf("%s%d", in.AppId, in.Timestamp), secretKey)
	if token != in.Token {
		return &services.MessageResponse{
			AppId:         in.AppId,
			ServiceId:     in.ServiceId,
			RespServiceId: "__CCS_ENTRY__",
			Timestamp:     time.Now().Unix(),
			Payload:       "__PAYLOAD__",
			Referers:      referers,
			Code:          20401, // Invalid Token
		}, nil
	}

	if in.NeedRespReferers {
		referers = append(referers, in.Referers...)
	}

	return &services.MessageResponse{
		AppId:         in.AppId,
		ServiceId:     in.ServiceId,
		RespServiceId: "__CCS_ENTRY__",
		Timestamp:     time.Now().Unix(),
		Payload:       "__PAYLOAD__",
		Referers:      referers,
		Code:          200,
	}, nil
}

type MessageServer struct {
	ListenAddr string
}

func NewMessageServer(listen string) *MessageServer {
	return &MessageServer{
		ListenAddr: listen,
	}
}

func (s *MessageServer) Start() {
	listen, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		log.Fatalf("Message Server failed to listen on %s: %v", s.ListenAddr, err)
	}

	grpcServer := grpc.NewServer()

	services.RegisterServiceMessageServer(grpcServer, &MessageService{})

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Message Server failed to serve: %v", err)
	}
}
