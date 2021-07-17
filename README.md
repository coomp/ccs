# CCS

## 平台角色

### 业务

+ BusinessId = 业务
  + AppId = 应用
    + ServiceId = 服务

+ Referer
  + TimeStamp
  + Value - SDK(CCS_SDK-1.0.1-10.0.0.12)/CCS(CCS_SVR-1.0.1-172.16.0.1) etc.

+ MessageConext
  + AppId - 应用 ID
  + ServiceId - 服务 ID
  + Token - 签名的 Token
  + TimeStamp - 消息产生的时间戳
  + Payload - 消息体
  + []Referer - 引用该消息的实体 [SDK -> CCS -> Service1 -> ServiceN] - 用于消息的跟踪和分析

## 一些逻辑

### 消息流转

+ 顺序消息
```
[SVC1 - OnReq] --> [CCS] --> [SVC2_REQ_MQ]
                                  |
                                  |
                                 \|/
                             [SVC2 - OnReq]
                                  /
                                 /
                               |//
                             [CCS] --> [SVC3_REQ_MQ]
                                            |
                                            |
                                           \|/
                                        [SVC3 - OnReq]
                                             /
                                            /
                                          |//
  [SVC1 - OnResp] <-- [SVC1_RESP_MQ] <-- [CCS]
```

+ 分支消息
```
[SVC1 - OnReq] --> [CCS] --> [SVC2_REQ_MQ]
                                      |
                                      |
                                     \|/
                                [SVC2 - OnReq]
                                     /
                                    /
                                  |//
                                [CCS] --> [SVC3_REQ_MQ]
                                  |            |
                                  |            |
                                 \|/           \
                              [SVC4_REQ_MQ]     \
                                  |             \|/
                                  |        [SVC3 - OnReq]
                                 \|/             /
                            [SVC4 - OnReq]      /
                                   \           /
                                    \         /
                                     \       /
                                      \     /
                                      \\| |//
[SVC1 - OnResp] <-- [SVC1_RESP_MQ] <-- [CCS]
```

### 基础的 Token 检验
+ Token = HmacSha256(AppId + TimeStamp, SecretKey)

## 准备工作

+ Download and install protoc [protoc](https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protoc-3.17.3-win64.zip)

+ protoc-gen-go
```bash
go get -u github.com/golang/protobuf/protoc-gen-go
```

+ grpc
```bash
# TODO You Know ...
go get google.golang.org/grpc
```

## 测试

+ Clone code
+ `go mod tidy`
+ `go build .`

## TODO List

- SDK 0%
  - TCP connect to server
- CCS 20%
  - Load configure and serve FSM - 40%
  - Handle messages from GRPC 0%
- Message Model 70%
- Tenant
  - A mq instance/cluster for a business
  - Each `ServiceId` has 2 `Topic`, include REQ / RESP