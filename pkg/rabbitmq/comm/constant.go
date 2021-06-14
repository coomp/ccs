package comm

type NodeState int

const (
	READY   NodeState = 0
	RUNNING NodeState = 1
	STOP    NodeState = 2
)

type Object struct {
	Value interface{}
}



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

type RpcRequest struct {
	_msgpack        struct{} `msgpack:",asArray"`
	Class           string   `json:"@type,omitempty" msgpack:"-"`
	TargetInterface string   `json:"targetInterface,omitempty"`
	Method          string   `json:"method,omitempty"`
	ArgTypes        []string `json:"argTypes,omitempty"`
	Args            []Object `json:"args,omitempty"`
}

type RpcResponse struct {
	_msgpack struct{} `msgpack:",asArray"`
	Class    string   `json:"@type,omitempty" msgpack:"-"`
	Data     *Object  `json:"data,omitempty"`
}
