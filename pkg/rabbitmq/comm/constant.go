package comm

import (
	"encoding/base64"
)

// NodeState TODO
type NodeState int

const (
	// READY TODO
	READY NodeState = 0
	// RUNNING TODO
	RUNNING NodeState = 1
	// STOP TODO
	STOP NodeState = 2
)

// Object TODO
type Object struct {
	Value interface{}
}

// Base64Byte TODO
type Base64Byte []byte

// UnmarshalText TODO
func (x *Base64Byte) UnmarshalText(text []byte) error {
	b, e := base64.StdEncoding.DecodeString(string(text))
	if e != nil {
		return e
	}
	*x = b
	return nil
}

// EncodeHeader TODO
func (msg *Message) EncodeHeader() error {
	if msg.Headers == nil {
		return nil
	}
	msg.Attribute = msg.Headers.Encode()
	return nil
}

// Message TODO
type Message struct {
	_msgpack  struct{}   `msgpack:",asArray"`
	Class     string     `json:"@type,omitempty" msgpack:"-"`
	Id        int64      `json:"id,omitempty"`
	Topic     string     `json:"topic,omitempty"`
	Data      Base64Byte `json:"data,omitempty"`
	Attribute string     `json:"attribute,omitempty"`
	Headers   Options    `json:"-" msgpack:"-"`
	Flag      int32      `json:"flag,omitempty"`
	End       bool       `json:"end,omitempty"`
	SourceIp  string     `json:"source_ip,omitempty"`
}

// HeartbeatResponse TODO
type HeartbeatResponse struct {
	_msgpack            struct{}                     `msgpack:",asArray"`
	Class               string                       `json:"@type,omitempty" msgpack:"-"`
	Code                int32                        `json:"code,omitempty"`
	Tips                string                       `json:"tips,omitempty"`
	Success             bool                         `json:"success,omitempty"`
	Consumers           []string                     `json:"consumers,omitempty"`
	TopicQueues         map[string][]Queue           `json:"topicQueues,omitempty"`
	AddQueueList        []Queue                      `json:"addQueueList,omitempty"`
	ModQueueList        []Queue                      `json:"modQueueList,omitempty"`
	Offset2Reset        map[string]map[string]string `json:"offset2Reset,omitempty"`
	ExtendAttributeMap  map[string]*Object           `json:"extendAttributeMap,omitempty"`
	ShareBrokerGroupMap map[string]BrokerGroup       `json:"shareBrokerGroupMap,omitempty"`
}

// RequestWrapper TODO
type RequestWrapper struct {
	_msgpack     struct{} `msgpack:",asArray" json:"__msgpack,omitempty"`
	Class        string   `json:"@type,omitempty" msgpack:"-"`
	Serial       int64    `json:"serial,omitempty"`
	CodecType    int32    `json:"codecType,omitempty"`
	ProtocolType int32    `json:"protocolType,omitempty"`
	RequestData  *Object  `json:"requestData,omitempty"`
	Timeout      int64    `json:"timeout,omitempty"`
	Compress     bool     `json:"compress,omitempty"`
}

// ResponseWrapper TODO
type ResponseWrapper struct {
	_msgpack     struct{} `msgpack:",asArray"`
	Class        string   `json:"@type,omitempty" msgpack:"-"`
	Serial       int64    `json:"serial,omitempty"`
	Success      bool     `json:"success,omitempty"`
	CodecType    int32    `json:"codecType,omitempty"`
	ProtocolType int32    `json:"protocolType,omitempty"`
	ResponseData *Object  `json:"responseData,omitempty"`
	ErrorMsg     string   `json:"errorMsg,omitempty"`
}

// RpcRequest TODO
type RpcRequest struct {
	_msgpack        struct{} `msgpack:",asArray"`
	Class           string   `json:"@type,omitempty" msgpack:"-"`
	TargetInterface string   `json:"targetInterface,omitempty"`
	Method          string   `json:"method,omitempty"`
	ArgTypes        []string `json:"argTypes,omitempty"`
	Args            []Object `json:"args,omitempty"`
}

// RpcResponse TODO
type RpcResponse struct {
	_msgpack struct{} `msgpack:",asArray"`
	Class    string   `json:"@type,omitempty" msgpack:"-"`
	Data     *Object  `json:"data,omitempty"`
}

// Broker TODO
type Broker struct {
	_msgpack struct{} `msgpack:",asArray"`
	Class    string   `json:"@type,omitempty" msgpack:"-"`
	Id       int32    `json:"id,omitempty"`
	Host     string   `json:"host,omitempty"`
	Port     int32    `json:"port,omitempty"`
	Type     int32    `json:"type,omitempty"`
	Name     string   `json:"name,omitempty"`
	IsActive bool     `json:"isActive,omitempty"`
}

// BrokerGroup TODO
type BrokerGroup struct {
	_msgpack       struct{} `msgpack:",asArray"`
	Class          string   `json:"@type,omitempty" msgpack:"-"`
	GroupName      string   `json:"groupName,omitempty"`
	Brokers        []Broker `json:"brokers,omitempty"`
	TimeStamp      int64    `json:"timeStamp,omitempty"`
	IsWritable     bool     `json:"isWritable,omitempty"`
	IsReadable     bool     `json:"isReadable,omitempty"`
	DelayTimeStamp int64    `json:"delayTimeStamp,omitempty"`
}

// HeartbeatRequest TODO
type HeartbeatRequest struct {
	_msgpack           struct{}               `msgpack:",asArray"`
	Class              string                 `json:"@type,omitempty" msgpack:"-"`
	ClientId           string                 `json:"clientId,omitempty"`
	Group              string                 `json:"group,omitempty"`
	Type               int32                  `json:"type,omitempty"`
	Version            int32                  `json:"version,omitempty"`
	Topics             []string               `json:"topics,omitempty"`
	BrokerGroup        *BrokerGroup           `json:"brokerGroup,omitempty"`
	TopicQueues        map[string][]Queue     `json:"topicQueues,omitempty"`
	BGInfoCollector    map[string]interface{} `json:"bGinfoCollector,omitempty"`
	ExtendAttributeMap map[string]*Object     `json:"extendAttributeMap,omitempty"`
}

// Queue TODO
type Queue struct {
	_msgpack       struct{}     `msgpack:",asArray"`
	Class          string       `json:"@type,omitempty" msgpack:"-"`
	Id             int64        `json:"id,omitempty"`
	BrokerGroup    *BrokerGroup `json:"brokerGroup,omitempty"`
	Topic          string       `json:"topic,omitempty"`
	Model          string       `json:"model,omitempty"`
	Writable       bool         `json:"writable,omitempty"`
	Readable       bool         `json:"readable,omitempty"`
	BrokerGroupKey int32        `json:"brokerGroupKey,omitempty"`
	Status         int32        `json:"status,omitempty"`
}
