package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/coomp/ccs/model"
	"github.com/coomp/ccs/pkg/repositories/rocketmq_impl"
	"github.com/coomp/ccs/pkg/services"
	"github.com/coomp/ccs/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MessageService struct {
	messageQueueService *services.MessageQueueService
	db                  *gorm.DB
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
			Code:          ERROR_INVALID_TOKEN, // Invalid Token
		}, nil
	}

	if in.NeedRespReferers {
		referers = append(referers, in.Referers...)
	}
	referers = append(referers, &services.Referer{
		Timestamp: time.Now().Unix(),
		Value:     "CCS_SERVER_1.0.0_127.0.0.1",
	})

	// TODO find next svc and generate topic from it

	var ccss model.CcsService
	svc.db.First(&ccss, "service_id = ?", in.ServiceId) // find product with code D42
	if ccss.ServiceId != in.ServiceId {
		// not found
		return &services.MessageResponse{
			AppId:         in.AppId,
			ServiceId:     in.ServiceId,
			RespServiceId: "__CCS_ENTRY__",
			Timestamp:     time.Now().Unix(),
			Payload:       "__PAYLOAD__",
			Referers:      referers,
			Code:          ERROR_SERVICE_NOT_FOUND,
		}, nil
	}

	topic := fmt.Sprintf("%s_%s_%s", in.AppId, in.ServiceId, "REQ")
	bytes, _ := json.Marshal(in)
	err := svc.messageQueueService.Send(topic, string(bytes))
	if err != nil {
		log.Printf("Error send message to message queue %v\n", err)
		return &services.MessageResponse{
			AppId:         in.AppId,
			ServiceId:     in.ServiceId,
			RespServiceId: "__CCS_ENTRY__",
			Timestamp:     time.Now().Unix(),
			Payload:       "__PAYLOAD__",
			Referers:      referers,
			Code:          ERROR_INTERNAL_SERVER_ERROR,
		}, nil
	}

	return &services.MessageResponse{
		AppId:         in.AppId,
		ServiceId:     in.ServiceId,
		RespServiceId: "__CCS_ENTRY__",
		Timestamp:     time.Now().Unix(),
		Payload:       "__PAYLOAD__",
		Referers:      referers,
		Code:          ERROR_SUCCESS,
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

func getDb() *gorm.DB {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Home dir:", u.HomeDir)

	os.MkdirAll(path.Join(u.HomeDir, ".ccsdb"), 0755)

	db, err := gorm.Open(sqlite.Open(path.Join(u.HomeDir, ".ccsdb", "sqlite.db")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func (s *MessageServer) Start() {
	listen, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		log.Fatalf("Message Server failed to listen on %s: %v", s.ListenAddr, err)
	}

	grpcServer := grpc.NewServer()

	messageQueueRepo, _ := rocketmq_impl.NewRocketMQMessageQueueRepository()

	messageService := &MessageService{
		messageQueueService: &services.MessageQueueService{Repo: messageQueueRepo},
		db:                  getDb(),
	}

	services.RegisterServiceMessageServer(grpcServer, messageService)

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Message Server failed to serve: %v", err)
	}
}
