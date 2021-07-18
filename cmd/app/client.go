package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/coomp/ccs/pkg/services"
	"github.com/coomp/ccs/util"
	"google.golang.org/grpc"
)

type MessageClient struct {
	Address string
}

func NewMessageClient(addr string) *MessageClient {
	return &MessageClient{
		Address: addr,
	}
}

func (cli *MessageClient) Echo(str string) {
	conn, err := grpc.Dial(cli.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	client := services.NewServiceMessageClient(conn)

	echoResponse, err := client.OnEcho(ctx, &services.EchoRequest{Str: str})
	if err != nil {
		log.Printf("Error call OnEcho: %v", err)
	}

	log.Printf("Call OnEcho success: %v", echoResponse.Str)
}

func (cli *MessageClient) MessageRequest() {
	conn, err := grpc.Dial(cli.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	client := services.NewServiceMessageClient(conn)

	secretKey := "__SECRET_KEY__"

	param := &services.MessageRequest{
		AppId:            "__APP_ID_",
		ServiceId:        "__SERVICE_ID__",
		Timestamp:        time.Now().Unix(),
		Payload:          "__PAYLOAD__",
		NeedRespReferers: true,
	}

	// Token = HmacSha256(AppId + TimeStamp, SecretKey)
	param.Token = util.HmacSha256Base64(fmt.Sprintf("%s%d", param.AppId, param.Timestamp), secretKey)

	echoResponse, err := client.OnMessageRequest(ctx, param)
	if err != nil {
		log.Printf("Error call OnEcho: %v", err)
	}

	bytes, _ := json.Marshal(echoResponse)
	log.Printf("Call OnMessageRequest success: %v", string(bytes))
}
