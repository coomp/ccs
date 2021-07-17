@REM generate code
protoc -I. --go_out=plugins=grpc:. ./grpc/message.proto